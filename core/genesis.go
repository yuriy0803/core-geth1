// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/params/confp"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/trie"
)

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

// ChainOverrides contains the changes to chain config.
type ChainOverrides struct {
	OverrideShanghai *uint64
	OverrideCancun   *uint64
	OverrideVerkle   *uint64
}

func ReadGenesis(db ethdb.Database) (*genesisT.Genesis, error) {
	var genesis genesisT.Genesis
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		return nil, fmt.Errorf("invalid genesis hash in database: %x", stored)
	}
	blob := rawdb.ReadGenesisStateSpec(db, stored)
	if blob == nil {
		return nil, errors.New("genesis state missing from db")
	}
	if len(blob) != 0 {
		if err := genesis.Alloc.UnmarshalJSON(blob); err != nil {
			return nil, fmt.Errorf("could not unmarshal genesis state json: %s", err)
		}
	}
	genesis.Config = rawdb.ReadChainConfig(db, stored)
	if genesis.Config == nil {
		return nil, errors.New("genesis config missing from db")
	}
	genesisBlock := rawdb.ReadBlock(db, stored, 0)
	if genesisBlock == nil {
		return nil, errors.New("genesis block missing from db")
	}
	genesisHeader := genesisBlock.Header()
	genesis.Nonce = genesisHeader.Nonce.Uint64()
	genesis.Timestamp = genesisHeader.Time
	genesis.ExtraData = genesisHeader.Extra
	genesis.GasLimit = genesisHeader.GasLimit
	genesis.Difficulty = genesisHeader.Difficulty
	genesis.Mixhash = genesisHeader.MixDigest
	genesis.Coinbase = genesisHeader.Coinbase
	genesis.BaseFee = genesisHeader.BaseFee
	genesis.ExcessBlobGas = genesisHeader.ExcessBlobGas
	genesis.BlobGasUsed = genesisHeader.BlobGasUsed

	return &genesis, nil
}

// SetupGenesisBlock wraps SetupGenesisBlockWithOverride, always using a nil value for the override.
func SetupGenesisBlock(db ethdb.Database, triedb *trie.Database, genesis *genesisT.Genesis) (ctypes.ChainConfigurator, common.Hash, error) {
	return SetupGenesisBlockWithOverride(db, triedb, genesis, nil)
}

func SetupGenesisBlockWithOverride(db ethdb.Database, triedb *trie.Database, genesis *genesisT.Genesis, overrides *ChainOverrides) (ctypes.ChainConfigurator, common.Hash, error) {
	if genesis != nil && confp.IsEmpty(genesis.Config) {
		return params.AllEthashProtocolChanges, common.Hash{}, genesisT.ErrGenesisNoConfig
	}

	applyOverrides := func(config ctypes.ChainConfigurator) {
		if config != nil {
			// Block-based overrides are not provided because Shanghai is
			// ETH-network specific and that protocol is defined exclusively in time-based forks.
			if overrides != nil && overrides.OverrideShanghai != nil {
				config.SetEIP3651TransitionTime(overrides.OverrideShanghai)
				config.SetEIP3855TransitionTime(overrides.OverrideShanghai)
				config.SetEIP3860TransitionTime(overrides.OverrideShanghai)
				config.SetEIP4895TransitionTime(overrides.OverrideShanghai)
				config.SetEIP6049TransitionTime(overrides.OverrideShanghai)
			}
			if overrides != nil && overrides.OverrideCancun != nil {
				config.SetEIP4844TransitionTime(overrides.OverrideCancun)
			}
			if overrides != nil && overrides.OverrideVerkle != nil {
				log.Warn("Verkle-fork is not yet supported")
			}
		}
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			log.Warn("Not specifying a chain flag is deprecated and will be removed in the future, please use --mainnet for Ethereum mainnet")
			genesis = params.DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}

		block, err := CommitGenesis(genesis, db, triedb)
		if err != nil {
			return genesis.Config, common.Hash{}, err
		}
		applyOverrides(genesis.Config)
		log.Info("Wrote genesis block OK", "config", genesis.Config)
		return genesis.Config, block.Hash(), nil
	}
	// We have the genesis block in database(perhaps in ancient database)
	// but the corresponding state is missing.
	header := rawdb.ReadHeader(db, stored, 0)
	if header.Root != types.EmptyRootHash && !rawdb.HasLegacyTrieNode(db, header.Root) {
		if genesis == nil {
			genesis = params.DefaultGenesisBlock()
		}
		// Ensure the stored genesis matches with the given one.
		hash := GenesisToBlock(genesis, nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &genesisT.GenesisMismatchError{Stored: stored, New: hash}
		}
		block, err := CommitGenesis(genesis, db, triedb)
		if err != nil {
			return genesis.Config, hash, err
		}
		applyOverrides(genesis.Config)
		return genesis.Config, block.Hash(), nil
	}
	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := GenesisToBlock(genesis, nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &genesisT.GenesisMismatchError{Stored: stored, New: hash}
		}
	}
	// Get the existing chain configuration.
	newcfg := configOrDefault(genesis, stored)
	applyOverrides(newcfg)
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)
		return newcfg, stored, nil
	} else {
		log.Info("Found stored genesis block", "config", storedcfg)
	}
	storedData, _ := json.Marshal(storedcfg)

	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	//
	// (meowsbits): The idea here is to use stored configs when they are not upgrade-able via defaults.
	// Pre-existing logic only upgraded mainnet, when it should upgrade all defaulty chains.
	// New logic (below) checks _inequality_ between a defaulty config and a stored config. If different,
	// the stored config is used. This breaks auto-upgrade magic for defaulty chains.
	if genesis == nil && !confp.Identical(storedcfg, newcfg, []string{"NetworkID", "ChainID"}) {
		// TODO/meowsbits/20220405: ethereum code for this scope follows:
		/*
				// Special case: if a private network is being used (no genesis and also no
				// mainnet hash in the database), we must not apply the `configOrDefault`
				// chain config as that would be AllProtocolChanges (applying any new fork
				// on top of an existing private network genesis block). In that case, only
				// apply the overrides.

				if ... :
				newcfg = storedcfg
			if overrideGrayGlacier != nil {
				newcfg.GrayGlacierBlock = overrideGrayGlacier
				}
				if overrideTerminalTotalDifficulty != nil {
					newcfg.TerminalTotalDifficulty = overrideTerminalTotalDifficulty
				}
		*/
		// ... and this is ours:
		log.Info("Found non-defaulty stored config, using it.")
		newcfg = storedcfg
		applyOverrides(newcfg)
		return storedcfg, stored, nil
	}
	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	head := rawdb.ReadHeadHeader(db)
	if head == nil {
		return newcfg, stored, fmt.Errorf("missing head header")
	}
	compatErr := confp.Compatible(head.Number, &head.Time, storedcfg, newcfg)
	if compatErr != nil && ((head.Number.Uint64() != 0 && compatErr.RewindToBlock != 0) || (head.Time != 0 && compatErr.RewindToTime != 0)) {
		return newcfg, stored, compatErr
	}
	// Don't overwrite if the old is identical to the new
	if newData, _ := json.Marshal(newcfg); !bytes.Equal(storedData, newData) {
		rawdb.WriteChainConfig(db, stored, newcfg)
	}
	return newcfg, stored, nil
}

// LoadCliqueConfig loads the stored clique config if the chain config
// is already present in database, otherwise, return the config in the
// provided genesis specification. Note the returned clique config can
// be nil if we are not in the clique network.
func LoadCliqueConfig(db ethdb.Database, genesis *genesisT.Genesis) (*ctypes.CliqueConfig, error) {
	// Load the stored chain config from the database. It can be nil
	// in case the database is empty. Notably, we only care about the
	// chain config corresponds to the canonical chain.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if stored != (common.Hash{}) {
		storedcfg := rawdb.ReadChainConfig(db, stored)
		if storedcfg != nil {
			if storedcfg.GetConsensusEngineType() == ctypes.ConsensusEngineT_Clique {
				return &ctypes.CliqueConfig{
					Period: storedcfg.GetCliquePeriod(),
					Epoch:  storedcfg.GetCliqueEpoch(),
				}, nil
			}
		}
	}
	// Load the clique config from the provided genesis specification.
	if genesis != nil {
		// Reject invalid genesis spec without valid chain config
		if genesis.Config == nil {
			return nil, errGenesisNoConfig
		}
		// If the canonical genesis header is present, but the chain
		// config is missing(initialize the empty leveldb with an
		// external ancient chain segment), ensure the provided genesis
		// is matched.
		genesisBlock := MustCommitGenesis(rawdb.NewMemoryDatabase(), genesis)
		if stored != (common.Hash{}) && genesisBlock.Hash() != stored {
			return nil, &genesisT.GenesisMismatchError{Stored: stored, New: genesisBlock.Hash()}
		}
		if genesis.Config.GetConsensusEngineType() == ctypes.ConsensusEngineT_Clique {
			return &ctypes.CliqueConfig{
				Period: genesis.Config.GetCliquePeriod(),
				Epoch:  genesis.Config.GetCliqueEpoch(),
			}, nil
		}
	}
	// There is no stored chain config and no new config provided,
	// In this case the default chain config(mainnet) will be used,
	// namely ethash is the specified consensus engine, return nil.
	return nil, nil
}

func configOrDefault(g *genesisT.Genesis, ghash common.Hash) ctypes.ChainConfigurator {
	switch {
	case g != nil:
		return g.Config
	case ghash == params.MainnetGenesisHash:
		return params.MainnetChainConfig
	case ghash == params.GoerliGenesisHash:
		return params.GoerliChainConfig
	case ghash == params.MordorGenesisHash:
		return params.MordorChainConfig
	case ghash == params.SepoliaGenesisHash:
		return params.SepoliaChainConfig
	case ghash == params.MintMeGenesisHash:
		return params.MintMeChainConfig
	default:
		return params.AllEthashProtocolChanges
	}
}

// Flush adds allocated genesis accounts into a fresh new statedb and
// commit the state changes into the given database handler.
func gaFlush(ga *genesisT.GenesisAlloc, db ethdb.Database) error {
	statedb, err := state.New(common.Hash{}, state.NewDatabaseWithConfig(db, &trie.Config{Preimages: true}), nil)
	if err != nil {
		return err
	}
	for addr, account := range *ga {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root, err := statedb.Commit(0, false)
	if err != nil {
		return err
	}
	// Commit newly generated states into disk if it's not empty.
	if root != types.EmptyRootHash {
		err = statedb.Database().TrieDB().Commit(root, true)
		if err != nil {
			return err
		}
	}

	// Marshal the genesis state specification and persist.
	blob, err := json.Marshal(ga)
	if err != nil {
		return err
	}
	rawdb.WriteGenesisStateSpec(db, root, blob)
	return nil
}

// gaDeriveHash computes the state root according to the genesis specification.
func gaDeriveHash(ga *genesisT.GenesisAlloc) (common.Hash, error) {
	// Create an ephemeral in-memory database for computing hash,
	// all the derived states will be discarded to not pollute disk.
	db := state.NewDatabase(rawdb.NewMemoryDatabase())
	statedb, err := state.New(types.EmptyRootHash, db, nil)
	if err != nil {
		return common.Hash{}, err
	}
	for addr, account := range *ga {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	return statedb.Commit(0, false)
}

// Write writes the json marshaled genesis state into database
// with the given block hash as the unique identifier.
func gaWrite(ga *genesisT.GenesisAlloc, db ethdb.KeyValueWriter, hash common.Hash) error {
	blob, err := json.Marshal(ga)
	if err != nil {
		return err
	}
	rawdb.WriteGenesisStateSpec(db, hash, blob)
	return nil
}

// CommitGenesisState loads the stored genesis state with the given block
// hash and commits them into the given database handler.
func CommitGenesisState(db ethdb.Database, hash common.Hash) error {
	var alloc genesisT.GenesisAlloc
	blob := rawdb.ReadGenesisStateSpec(db, hash)
	if len(blob) != 0 {
		if err := alloc.UnmarshalJSON(blob); err != nil {
			return err
		}
	} else {
		// Genesis allocation is missing and there are several possibilities:
		// the node is legacy which doesn't persist the genesis allocation or
		// the persisted allocation is just lost.
		// - supported networks(mainnet, testnets), recover with defined allocations
		// - private network, can't recover
		var genesis *genesisT.Genesis
		switch hash {
		case params.MainnetGenesisHash:
			genesis = params.DefaultGenesisBlock()
		case params.GoerliGenesisHash:
			genesis = params.DefaultGoerliGenesisBlock()
		case params.SepoliaGenesisHash:
			genesis = params.DefaultSepoliaGenesisBlock()
		case params.MordorGenesisHash:
			genesis = params.DefaultMordorGenesisBlock()
		case params.MintMeGenesisHash:
			genesis = params.DefaultMintMeGenesisBlock()
		}
		if genesis != nil {
			alloc = genesis.Alloc
		} else {
			return errors.New("not found")
		}
	}
	err := gaFlush(&alloc, db)
	return err
}

// GenesisToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func GenesisToBlock(g *genesisT.Genesis, db ethdb.Database) *types.Block {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	root, err := gaDeriveHash(&g.Alloc)
	if err != nil {
		panic(err)
	}
	err = gaFlush(&g.Alloc, db)
	if err != nil {
		panic(err)
	}
	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       g.Timestamp,
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		BaseFee:    g.BaseFee,
		Difficulty: g.Difficulty,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = vars.GenesisGasLimit
	}
	// -- meowsbits/202203 go-ethereum has: if g.Difficulty == nil && g.Mixhash == (common.Hash{}) {
	// They also assign the Difficulty field directly.
	if g.Difficulty == nil {
		head.Difficulty = new(big.Int)
		head.Difficulty.Set(vars.GenesisDifficulty)
	}
	if g.Config != nil && g.Config.IsEnabled(g.Config.GetEIP1559Transition, common.Big0) {
		if g.BaseFee != nil {
			head.BaseFee = g.BaseFee
		} else {
			head.BaseFee = new(big.Int).SetUint64(vars.InitialBaseFee)
		}
	}
	var withdrawals []*types.Withdrawal
	if conf := g.Config; conf != nil {
		// EIP4895 defines the withdwrawals tx type, implemented on ETH in the Shanghai fork.
		isEIP4895 := conf.IsEnabledByTime(g.Config.GetEIP4895TransitionTime, &g.Timestamp) || g.Config.IsEnabled(g.Config.GetEIP4895Transition, new(big.Int).SetUint64(g.Number))
		if isEIP4895 {
			head.WithdrawalsHash = &types.EmptyWithdrawalsHash
			withdrawals = make([]*types.Withdrawal, 0)
		}
		isCancun := conf.IsEnabledByTime(g.Config.GetEIP4844TransitionTime, &g.Timestamp)
		if isCancun {
			head.ExcessBlobGas = g.ExcessBlobGas
			head.BlobGasUsed = g.BlobGasUsed
			if head.ExcessBlobGas == nil {
				head.ExcessBlobGas = new(uint64)
			}
			if head.BlobGasUsed == nil {
				head.BlobGasUsed = new(uint64)
			}
		}
	}
	return types.NewBlock(head, nil, nil, nil, trie.NewStackTrie(nil)).WithWithdrawals(withdrawals)
}

// CommitGenesis writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func CommitGenesis(g *genesisT.Genesis, db ethdb.Database, triedb *trie.Database) (*types.Block, error) {
	block := GenesisToBlock(g, db)
	if block.Number().Sign() != 0 {
		return nil, errors.New("can't commit genesis block with number > 0")
	}
	config := g.Config
	if config == nil {
		config = params.AllEthashProtocolChanges
	}

	// Upstream omission:
	// ethereum/go-ethereum does: config.CheckConfigForkOrder()
	// core-geth does not.

	if config.GetConsensusEngineType().IsClique() && len(block.Extra()) == 0 {
		return nil, errors.New("can't start clique chain without signers")
	}
	// All the checks has passed, flush the states derived from the genesis
	// specification as well as the specification itself into the provided
	// database.
	if err := gaWrite(&g.Alloc, db, block.Hash()); err != nil {
		return nil, err
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), block.Difficulty())
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadFastBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// MustCommitGenesis writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
// Note the state changes will be committed in hash-based scheme, use Commit
// if path-scheme is preferred.
func MustCommitGenesis(db ethdb.Database, g *genesisT.Genesis) *types.Block {
	block, err := CommitGenesis(g, db, trie.NewDatabase(db))
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db ethdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := genesisT.Genesis{
		Alloc:   genesisT.GenesisAlloc{addr: {Balance: balance}},
		BaseFee: big.NewInt(vars.InitialBaseFee),
	}
	return MustCommitGenesis(db, &g)
}
