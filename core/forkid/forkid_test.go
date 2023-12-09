// Copyright 2019 The go-ethereum Authors
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

package forkid

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/yuriy0803/core-geth1/common"
	"github.com/yuriy0803/core-geth1/params"
	"github.com/yuriy0803/core-geth1/params/confp"
	"github.com/yuriy0803/core-geth1/params/types/coregeth"
	"github.com/yuriy0803/core-geth1/params/types/ctypes"
	"github.com/yuriy0803/core-geth1/rlp"
)

// TestCreation tests that different genesis and fork rule combinations result in
// the correct fork ID.
func TestCreation(t *testing.T) {
	type testcase struct {
		head uint64
		time uint64
		want ID
	}
	tests := []struct {
		name    string
		config  ctypes.ChainConfigurator
		genesis common.Hash
		cases   []testcase
	}{
		// Mainnet test cases
		{
			"mainnet",
			params.MainnetChainConfig,
			params.MainnetGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},                    // Unsynced
				{1149999, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},              // Last Frontier block
				{1150000, 0, ID{Hash: checksumToBytes(0x97c2c34c), Next: 1920000}},              // First Homestead block
				{1919999, 0, ID{Hash: checksumToBytes(0x97c2c34c), Next: 1920000}},              // Last Homestead block
				{1920000, 0, ID{Hash: checksumToBytes(0x91d1f948), Next: 2463000}},              // First DAO block
				{2462999, 0, ID{Hash: checksumToBytes(0x91d1f948), Next: 2463000}},              // Last DAO block
				{2463000, 0, ID{Hash: checksumToBytes(0x7a64da13), Next: 2675000}},              // First Tangerine block
				{2674999, 0, ID{Hash: checksumToBytes(0x7a64da13), Next: 2675000}},              // Last Tangerine block
				{2675000, 0, ID{Hash: checksumToBytes(0x3edd5b10), Next: 4370000}},              // First Spurious block
				{4369999, 0, ID{Hash: checksumToBytes(0x3edd5b10), Next: 4370000}},              // Last Spurious block
				{4370000, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7280000}},              // First Byzantium block
				{7279999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7280000}},              // Last Byzantium block
				{7280000, 0, ID{Hash: checksumToBytes(0x668db0af), Next: 9069000}},              // First and last Constantinople, first Petersburg block
				{9068999, 0, ID{Hash: checksumToBytes(0x668db0af), Next: 9069000}},              // Last Petersburg block
				{9069000, 0, ID{Hash: checksumToBytes(0x879d6e30), Next: 9200000}},              // First Istanbul and first Muir Glacier block
				{9199999, 0, ID{Hash: checksumToBytes(0x879d6e30), Next: 9200000}},              // Last Istanbul and first Muir Glacier block
				{9200000, 0, ID{Hash: checksumToBytes(0xe029e991), Next: 12244000}},             // First Muir Glacier block
				{12243999, 0, ID{Hash: checksumToBytes(0xe029e991), Next: 12244000}},            // Last Muir Glacier block
				{12244000, 0, ID{Hash: checksumToBytes(0x0eb440f6), Next: 12965000}},            // First Berlin block
				{12964999, 0, ID{Hash: checksumToBytes(0x0eb440f6), Next: 12965000}},            // Last Berlin block
				{12965000, 0, ID{Hash: checksumToBytes(0xb715077d), Next: 13773000}},            // First London block
				{13772999, 0, ID{Hash: checksumToBytes(0xb715077d), Next: 13773000}},            // Last London block
				{13773000, 0, ID{Hash: checksumToBytes(0x20c327fc), Next: 15050000}},            // First Arrow Glacier block
				{15049999, 0, ID{Hash: checksumToBytes(0x20c327fc), Next: 15050000}},            // Last Arrow Glacier block
				{15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1681338455}},          // First Gray Glacier block
				{20000000, 1681338454, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1681338455}}, // Last Gray Glacier block
				{20000000, 1681338455, ID{Hash: checksumToBytes(0xdce96c2d), Next: 0}},          // First Shanghai block
				{30000000, 2000000000, ID{Hash: checksumToBytes(0xdce96c2d), Next: 0}},          // Future Shanghai block
			},
		},
		// Goerli test cases
		{
			"goerli",
			params.GoerliChainConfig,
			params.GoerliGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0xa3f5ab08), Next: 1561651}},                   // Unsynced, last Frontier, Homestead, Tangerine, Spurious, Byzantium, Constantinople and first Petersburg block
				{1561650, 0, ID{Hash: checksumToBytes(0xa3f5ab08), Next: 1561651}},             // Last Petersburg block
				{1561651, 0, ID{Hash: checksumToBytes(0xc25efa5c), Next: 4460644}},             // First Istanbul block
				{4460643, 0, ID{Hash: checksumToBytes(0xc25efa5c), Next: 4460644}},             // Last Istanbul block
				{4460644, 0, ID{Hash: checksumToBytes(0x757a1c47), Next: 5062605}},             // First Berlin block
				{5000000, 0, ID{Hash: checksumToBytes(0x757a1c47), Next: 5062605}},             // Last Berlin block
				{5062605, 0, ID{Hash: checksumToBytes(0xB8C6299D), Next: 1678832736}},          // First London block
				{6000000, 1678832735, ID{Hash: checksumToBytes(0xB8C6299D), Next: 1678832736}}, // Last London block
				{6000001, 1678832736, ID{Hash: checksumToBytes(0xf9843abf), Next: 0}},          // First Shanghai block
				{6500000, 2678832736, ID{Hash: checksumToBytes(0xf9843abf), Next: 0}},          // Future Shanghai block
			},
		},
		// Sepolia test cases
		{
			"sepolia",
			params.SepoliaChainConfig,
			params.SepoliaGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0xfe3366e7), Next: 1735371}},                   // Unsynced, last Frontier, Homestead, Tangerine, Spurious, Byzantium, Constantinople, Petersburg, Istanbul, Berlin and first London block
				{1735370, 0, ID{Hash: checksumToBytes(0xfe3366e7), Next: 1735371}},             // Last London block
				{1735371, 0, ID{Hash: checksumToBytes(0xb96cbd13), Next: 1677557088}},          // First MergeNetsplit block
				{1735372, 1677557087, ID{Hash: checksumToBytes(0xb96cbd13), Next: 1677557088}}, // Last MergeNetsplit block
				{1735372, 1677557088, ID{Hash: checksumToBytes(0xf7f9bc08), Next: 0}},          // First Shanghai block
			},
		},
		{
			"classic",
			params.ClassicChainConfig,
			params.MainnetGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{1, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{2, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{3, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{9, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{10, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{1149999, 0, ID{Hash: checksumToBytes(0xfc64ec04), Next: 1150000}},
				{1150000, 0, ID{Hash: checksumToBytes(0x97c2c34c), Next: 2500000}},
				{1150001, 0, ID{Hash: checksumToBytes(0x97c2c34c), Next: 2500000}},
				{2499999, 0, ID{Hash: checksumToBytes(0x97c2c34c), Next: 2500000}},
				{2500000, 0, ID{Hash: checksumToBytes(0xdb06803f), Next: 3000000}},
				{2500001, 0, ID{Hash: checksumToBytes(0xdb06803f), Next: 3000000}},
				{2999999, 0, ID{Hash: checksumToBytes(0xdb06803f), Next: 3000000}},
				{3000000, 0, ID{Hash: checksumToBytes(0xaff4bed4), Next: 5000000}},
				{3000001, 0, ID{Hash: checksumToBytes(0xaff4bed4), Next: 5000000}},
				{4999999, 0, ID{Hash: checksumToBytes(0xaff4bed4), Next: 5000000}},
				{5000000, 0, ID{Hash: checksumToBytes(0xf79a63c0), Next: 5900000}},
				{5000001, 0, ID{Hash: checksumToBytes(0xf79a63c0), Next: 5900000}},
				{5899999, 0, ID{Hash: checksumToBytes(0xf79a63c0), Next: 5900000}},
				{5900000, 0, ID{Hash: checksumToBytes(0x744899d6), Next: 8772000}},
				{5900001, 0, ID{Hash: checksumToBytes(0x744899d6), Next: 8772000}},
				{8771999, 0, ID{Hash: checksumToBytes(0x744899d6), Next: 8772000}},
				{8772000, 0, ID{Hash: checksumToBytes(0x518b59c6), Next: 9573000}},
				{8772001, 0, ID{Hash: checksumToBytes(0x518b59c6), Next: 9573000}},
				{9572999, 0, ID{Hash: checksumToBytes(0x518b59c6), Next: 9573000}},
				{9573000, 0, ID{Hash: checksumToBytes(0x7ba22882), Next: 10500839}},
				{9573001, 0, ID{Hash: checksumToBytes(0x7ba22882), Next: 10500839}},
				{10500838, 0, ID{Hash: checksumToBytes(0x7ba22882), Next: 10500839}},
				{10500839, 0, ID{Hash: checksumToBytes(0x9007bfcc), Next: 11_700_000}},
				{10500840, 0, ID{Hash: checksumToBytes(0x9007bfcc), Next: 11_700_000}},
				{11_699_999, 0, ID{Hash: checksumToBytes(0x9007bfcc), Next: 11_700_000}},
				{11_700_000, 0, ID{Hash: checksumToBytes(0xdb63a1ca), Next: 13_189_133}},
				{11_700_001, 0, ID{Hash: checksumToBytes(0xdb63a1ca), Next: 13_189_133}},
				{13_189_132, 0, ID{Hash: checksumToBytes(0xdb63a1ca), Next: 13_189_133}},
				{13_189_133, 0, ID{Hash: checksumToBytes(0x0f6bf187), Next: 14_525_000}},
				{13_189_134, 0, ID{Hash: checksumToBytes(0x0f6bf187), Next: 14_525_000}},
				{14_524_999, 0, ID{Hash: checksumToBytes(0x0f6bf187), Next: 14_525_000}},
				{14_525_000, 0, ID{Hash: checksumToBytes(0x7fd1bb25), Next: 19_250_000}},
				{14_525_001, 0, ID{Hash: checksumToBytes(0x7fd1bb25), Next: 19_250_000}},
				{19_249_999, 0, ID{Hash: checksumToBytes(0x7fd1bb25), Next: 19_250_000}},
				{19_250_000, 0, ID{Hash: checksumToBytes(0xbe46d57c), Next: 0}},
				{19_250_001, 0, ID{Hash: checksumToBytes(0xbe46d57c), Next: 0}},
			},
		},
		{
			"mordor",
			params.MordorChainConfig,
			params.MordorGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{1, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{2, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{3, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{9, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{10, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{301242, 0, ID{Hash: checksumToBytes(0x175782aa), Next: 301243}},
				{301243, 0, ID{Hash: checksumToBytes(0x604f6ee1), Next: 999983}},
				{301244, 0, ID{Hash: checksumToBytes(0x604f6ee1), Next: 999983}},
				{999982, 0, ID{Hash: checksumToBytes(0x604f6ee1), Next: 999983}},
				{999983, 0, ID{Hash: checksumToBytes(0xf42f5539), Next: 2_520_000}},
				{999984, 0, ID{Hash: checksumToBytes(0xf42f5539), Next: 2_520_000}},
				{2_519_999, 0, ID{Hash: checksumToBytes(0xf42f5539), Next: 2_520_000}},
				{2_520_000, 0, ID{Hash: checksumToBytes(0x66b5c286), Next: 3_985_893}},
				{3_985_892, 0, ID{Hash: checksumToBytes(0x66b5c286), Next: 3_985_893}},
				{3_985_893, 0, ID{Hash: checksumToBytes(0x92b323e0), Next: 5_520_000}},
				{3_985_894, 0, ID{Hash: checksumToBytes(0x92b323e0), Next: 5_520_000}},
				{5_519_999, 0, ID{Hash: checksumToBytes(0x92b323e0), Next: 5_520_000}},
				{5_520_000, 0, ID{Hash: checksumToBytes(0x8c9b1797), Next: 9_957_000}},
				{5_520_001, 0, ID{Hash: checksumToBytes(0x8c9b1797), Next: 9_957_000}},
				{9_956_999, 0, ID{Hash: checksumToBytes(0x8c9b1797), Next: 9_957_000}},
				{9_957_000, 0, ID{Hash: checksumToBytes(0x3a6b00d7), Next: 0}},
				{9_957_001, 0, ID{Hash: checksumToBytes(0x3a6b00d7), Next: 0}},
			},
		},
		// MintMe test cases
		{
			"mintme",
			params.MintMeChainConfig,
			params.MintMeGenesisHash,
			[]testcase{
				{0, 0, ID{Hash: checksumToBytes(0x02bf4180), Next: 252500}},
				{252500, 0, ID{Hash: checksumToBytes(0x50aed09f), Next: 0}},
			},
		},
	}
	for i, tt := range tests {
		for j, ttt := range tt.cases {
			if have := NewID(tt.config, tt.genesis, ttt.head, ttt.time); have != ttt.want {
				t.Errorf("test %d, case %d: fork ID mismatch: have %x, want %x", i, j, have, ttt.want)
			}
		}
	}
}

// TestValidation tests that a local peer correctly validates and accepts a remote
// fork ID.
func TestValidation(t *testing.T) {
	// Config that has not timestamp enabled
	legacyConfig := *params.MainnetChainConfig
	legacyConfig.ShanghaiTime = nil

	tests := []struct {
		config ctypes.ChainConfigurator
		head   uint64
		time   uint64
		id     ID
		err    error
	}{
		// ------------------
		// Block based tests
		// ------------------

		// Local is mainnet Gray Glacier, remote announces the same. No future fork is announced.
		{&legacyConfig, 15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 0}, nil},

		// Local is mainnet Gray Glacier, remote announces the same. Remote also announces a next fork
		// at block 0xffffffff, but that is uncertain.
		{&legacyConfig, 15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: math.MaxUint64}, nil},

		// Local is mainnet currently in Byzantium only (so it's aware of Petersburg), remote announces
		// also Byzantium, but it's not yet aware of Petersburg (e.g. non updated node before the fork).
		// In this case we don't know if Petersburg passed yet or not.
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 0}, nil},

		// Local is mainnet currently in Byzantium only (so it's aware of Petersburg), remote announces
		// also Byzantium, and it's also aware of Petersburg (e.g. updated node before the fork). We
		// don't know if Petersburg passed yet (will pass) or not.
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7280000}, nil},

		// Local is mainnet currently in Byzantium only (so it's aware of Petersburg), remote announces
		// also Byzantium, and it's also aware of some random fork (e.g. misconfigured Petersburg). As
		// neither forks passed at neither nodes, they may mismatch, but we still connect for now.
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: math.MaxUint64}, nil},

		// Local is mainnet exactly on Petersburg, remote announces Byzantium + knowledge about Petersburg. Remote
		// is simply out of sync, accept.
		{&legacyConfig, 7280000, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7280000}, nil},

		// Local is mainnet Petersburg, remote announces Byzantium + knowledge about Petersburg. Remote
		// is simply out of sync, accept.
		{&legacyConfig, 7987396, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7280000}, nil},

		// Local is mainnet Petersburg, remote announces Spurious + knowledge about Byzantium. Remote
		// is definitely out of sync. It may or may not need the Petersburg update, we don't know yet.
		{&legacyConfig, 7987396, 0, ID{Hash: checksumToBytes(0x3edd5b10), Next: 4370000}, nil},

		// Local is mainnet Byzantium, remote announces Petersburg. Local is out of sync, accept.
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0x668db0af), Next: 0}, nil},

		// Local is mainnet Spurious, remote announces Byzantium, but is not aware of Petersburg. Local
		// out of sync. Local also knows about a future fork, but that is uncertain yet.
		{&legacyConfig, 4369999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 0}, nil},

		// Local is mainnet Petersburg. remote announces Byzantium but is not aware of further forks.
		// Remote needs software update.
		{&legacyConfig, 7987396, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 0}, ErrRemoteStale},

		// Local is mainnet Petersburg, and isn't aware of more forks. Remote announces Petersburg +
		// 0xffffffff. Local needs software update, reject.
		{&legacyConfig, 7987396, 0, ID{Hash: checksumToBytes(0x5cddc0e1), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Byzantium, and is aware of Petersburg. Remote announces Petersburg +
		// 0xffffffff. Local needs software update, reject.
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0x5cddc0e1), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Petersburg, remote is Rinkeby Petersburg.
		{&legacyConfig, 7987396, 0, ID{Hash: checksumToBytes(0xafec6b27), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Gray Glacier, far in the future. Remote announces Gopherium (non existing fork)
		// at some future block 88888888, for itself, but past block for local. Local is incompatible.
		//
		// This case detects non-upgraded nodes with majority hash power (typical Ropsten mess).
		//
		// TODO(karalabe): This testcase will fail once mainnet gets timestamped forks, make legacy chain config
		{&legacyConfig, 88888888, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 88888888}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Byzantium. Remote is also in Byzantium, but announces Gopherium (non existing
		// fork) at block 7279999, before Petersburg. Local is incompatible.
		//
		// TODO(karalabe): This testcase will fail once mainnet gets timestamped forks, make legacy chain config
		{&legacyConfig, 7279999, 0, ID{Hash: checksumToBytes(0xa00bc324), Next: 7279999}, ErrLocalIncompatibleOrStale},

		// ------------------------------------
		// Block to timestamp transition tests
		// ------------------------------------

		// Local is mainnet currently in Gray Glacier only (so it's aware of Shanghai), remote announces
		// also Gray Glacier, but it's not yet aware of Shanghai (e.g. non updated node before the fork).
		// In this case we don't know if Shanghai passed yet or not.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 0}, nil},

		// Local is mainnet currently in Gray Glacier only (so it's aware of Shanghai), remote announces
		// also Gray Glacier, and it's also aware of Shanghai (e.g. updated node before the fork). We
		// don't know if Shanghai passed yet (will pass) or not.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1681338455}, nil},

		// Local is mainnet currently in Gray Glacier only (so it's aware of Shanghai), remote announces
		// also Gray Glacier, and it's also aware of some random fork (e.g. misconfigured Shanghai). As
		// neither forks passed at neither nodes, they may mismatch, but we still connect for now.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: math.MaxUint64}, nil},

		// Local is mainnet exactly on Shanghai, remote announces Gray Glacier + knowledge about Shanghai. Remote
		// is simply out of sync, accept.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1681338455}, nil},

		// Local is mainnet Shanghai, remote announces Gray Glacier + knowledge about Shanghai. Remote
		// is simply out of sync, accept.
		{params.MainnetChainConfig, 20123456, 1681338456, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1681338455}, nil},

		// Local is mainnet Shanghai, remote announces Arrow Glacier + knowledge about Gray Glacier. Remote
		// is definitely out of sync. It may or may not need the Shanghai update, we don't know yet.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0x20c327fc), Next: 15050000}, nil},

		// Local is mainnet Gray Glacier, remote announces Shanghai. Local is out of sync, accept.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(0xdce96c2d), Next: 0}, nil},

		// Local is mainnet Arrow Glacier, remote announces Gray Glacier, but is not aware of Shanghai. Local
		// out of sync. Local also knows about a future fork, but that is uncertain yet.
		{params.MainnetChainConfig, 13773000, 0, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 0}, nil},

		// Local is mainnet Shanghai. remote announces Gray Glacier but is not aware of further forks.
		// Remote needs software update.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 0}, ErrRemoteStale},

		// Local is mainnet Gray Glacier, and isn't aware of more forks. Remote announces Gray Glacier +
		// 0xffffffff. Local needs software update, reject.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(checksumUpdate(0xf0afd0e3, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Gray Glacier, and is aware of Shanghai. Remote announces Shanghai +
		// 0xffffffff. Local needs software update, reject.
		{params.MainnetChainConfig, 15050000, 0, ID{Hash: checksumToBytes(checksumUpdate(0xdce96c2d, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Gray Glacier, far in the future. Remote announces Gopherium (non existing fork)
		// at some future timestamp 8888888888, for itself, but past block for local. Local is incompatible.
		//
		// This case detects non-upgraded nodes with majority hash power (typical Ropsten mess).
		{params.MainnetChainConfig, 888888888, 1660000000, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1660000000}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Gray Glacier. Remote is also in Gray Glacier, but announces Gopherium (non existing
		// fork) at block 7279999, before Shanghai. Local is incompatible.
		{params.MainnetChainConfig, 19999999, 1667999999, ID{Hash: checksumToBytes(0xf0afd0e3), Next: 1667999999}, ErrLocalIncompatibleOrStale},

		// ----------------------
		// Timestamp based tests
		// ----------------------

		// Local is mainnet Shanghai, remote announces the same. No future fork is announced.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0xdce96c2d), Next: 0}, nil},

		// Local is mainnet Shanghai, remote announces the same. Remote also announces a next fork
		// at time 0xffffffff, but that is uncertain.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0xdce96c2d), Next: math.MaxUint64}, nil},

		// Local is mainnet currently in Shanghai only (so it's aware of Cancun), remote announces
		// also Shanghai, but it's not yet aware of Cancun (e.g. non updated node before the fork).
		// In this case we don't know if Cancun passed yet or not.
		//
		// TODO(karalabe): Enable this when Cancun is specced
		// {params.MainnetChainConfig, 20000000, 1668000000, ID{Hash: checksumToBytes(0x71147644), Next: 0}, nil},

		// Local is mainnet currently in Shanghai only (so it's aware of Cancun), remote announces
		// also Shanghai, and it's also aware of Cancun (e.g. updated node before the fork). We
		// don't know if Cancun passed yet (will pass) or not.
		//
		// TODO(karalabe): Enable this when Cancun is specced and update next timestamp
		// {params.MainnetChainConfig, 20000000, 1668000000, ID{Hash: checksumToBytes(0x71147644), Next: 1678000000}, nil},

		// Local is mainnet currently in Shanghai only (so it's aware of Cancun), remote announces
		// also Shanghai, and it's also aware of some random fork (e.g. misconfigured Cancun). As
		// neither forks passed at neither nodes, they may mismatch, but we still connect for now.
		//
		// TODO(karalabe): Enable this when Cancun is specced
		// {params.MainnetChainConfig, 20000000, 1668000000, ID{Hash: checksumToBytes(0x71147644), Next: math.MaxUint64}, nil},

		// Local is mainnet exactly on Cancun, remote announces Shanghai + knowledge about Cancun. Remote
		// is simply out of sync, accept.
		//
		// TODO(karalabe): Enable this when Cancun is specced, update local head and time, next timestamp
		// {params.MainnetChainConfig, 21000000, 1678000000, ID{Hash: checksumToBytes(0x71147644), Next: 1678000000}, nil},

		// Local is mainnet Cancun, remote announces Shanghai + knowledge about Cancun. Remote
		// is simply out of sync, accept.
		// TODO(karalabe): Enable this when Cancun is specced, update local head and time, next timestamp
		// {params.MainnetChainConfig, 21123456, 1678123456, ID{Hash: checksumToBytes(0x71147644), Next: 1678000000}, nil},

		// Local is mainnet Prague, remote announces Shanghai + knowledge about Cancun. Remote
		// is definitely out of sync. It may or may not need the Prague update, we don't know yet.
		//
		// TODO(karalabe): Enable this when Cancun **and** Prague is specced, update all the numbers
		// {params.MainnetChainConfig, 0, 0, ID{Hash: checksumToBytes(0x3edd5b10), Next: 4370000}, nil},

		// Local is mainnet Shanghai, remote announces Cancun. Local is out of sync, accept.
		//
		// TODO(karalabe): Enable this when Cancun is specced, update remote checksum
		// {params.MainnetChainConfig, 21000000, 1678000000, ID{Hash: checksumToBytes(0x00000000), Next: 0}, nil},

		// Local is mainnet Shanghai, remote announces Cancun, but is not aware of Prague. Local
		// out of sync. Local also knows about a future fork, but that is uncertain yet.
		//
		// TODO(karalabe): Enable this when Cancun **and** Prague is specced, update remote checksum
		// {params.MainnetChainConfig, 21000000, 1678000000, ID{Hash: checksumToBytes(0x00000000), Next: 0}, nil},

		// Local is mainnet Cancun. remote announces Shanghai but is not aware of further forks.
		// Remote needs software update.
		//
		// TODO(karalabe): Enable this when Cancun is specced, update local head and time
		// {params.MainnetChainConfig, 21000000, 1678000000, ID{Hash: checksumToBytes(0x71147644), Next: 0}, ErrRemoteStale},

		// Local is mainnet Shanghai, and isn't aware of more forks. Remote announces Shanghai +
		// 0xffffffff. Local needs software update, reject.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(checksumUpdate(0xdce96c2d, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Shanghai, and is aware of Cancun. Remote announces Cancun +
		// 0xffffffff. Local needs software update, reject.
		//
		// TODO(karalabe): Enable this when Cancun is specced, update remote checksum
		// {params.MainnetChainConfig, 20000000, 1668000000, ID{Hash: checksumToBytes(checksumUpdate(0x00000000, math.MaxUint64)), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Shanghai, remote is random Shanghai.
		{params.MainnetChainConfig, 20000000, 1681338455, ID{Hash: checksumToBytes(0x12345678), Next: 0}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Shanghai, far in the future. Remote announces Gopherium (non existing fork)
		// at some future timestamp 8888888888, for itself, but past block for local. Local is incompatible.
		//
		// This case detects non-upgraded nodes with majority hash power (typical Ropsten mess).
		{params.MainnetChainConfig, 88888888, 8888888888, ID{Hash: checksumToBytes(0xdce96c2d), Next: 8888888888}, ErrLocalIncompatibleOrStale},

		// Local is mainnet Shanghai. Remote is also in Shanghai, but announces Gopherium (non existing
		// fork) at timestamp 1668000000, before Cancun. Local is incompatible.
		//
		// TODO(karalabe): Enable this when Cancun is specced
		// {params.MainnetChainConfig, 20999999, 1677999999, ID{Hash: checksumToBytes(0x71147644), Next: 1678000000}, ErrLocalIncompatibleOrStale},
	}
	for i, tt := range tests {
		filter := newFilter(tt.config, params.MainnetGenesisHash, func() (uint64, uint64) { return tt.head, tt.time })
		if err := filter(tt.id); err != tt.err {
			t.Errorf("test %d, head: %d: validation error mismatch: have %v, want %v\nConfig=%s", i, tt.head, err, tt.err, tt.config)
		}
	}
}

// Tests that IDs are properly RLP encoded (specifically important because we
// use uint32 to store the hash, but we need to encode it as [4]byte).
func TestEncoding(t *testing.T) {
	tests := []struct {
		id   ID
		want []byte
	}{
		{ID{Hash: checksumToBytes(0), Next: 0}, common.Hex2Bytes("c6840000000080")},
		{ID{Hash: checksumToBytes(0xdeadbeef), Next: 0xBADDCAFE}, common.Hex2Bytes("ca84deadbeef84baddcafe,")},
		{ID{Hash: checksumToBytes(math.MaxUint32), Next: math.MaxUint64}, common.Hex2Bytes("ce84ffffffff88ffffffffffffffff")},
	}
	for i, tt := range tests {
		have, err := rlp.EncodeToBytes(tt.id)
		if err != nil {
			t.Errorf("test %d: failed to encode forkid: %v", i, err)
			continue
		}
		if !bytes.Equal(have, tt.want) {
			t.Errorf("test %d: RLP mismatch: have %x, want %x", i, have, tt.want)
		}
	}
}

func TestGatherForks(t *testing.T) {
	cases := []struct {
		name       string
		config     ctypes.ChainConfigurator
		blockForks []uint64
		timeForks  []uint64
	}{
		{
			"classic",
			params.ClassicChainConfig,
			[]uint64{1150000, 2500000, 3000000, 5000000, 5900000, 8772000, 9573000, 10500839, 11_700_000, 13_189_133, 14_525_000, 19_250_000},
			[]uint64{},
		},
		{
			"mainnet",
			params.MainnetChainConfig,
			[]uint64{1150000, 1920000, 2463000, 2675000, 4370000, 7280000, 9069000, 9200000, 12_244_000, 12_965_000, 13_773_000, 15050000},
			[]uint64{1681338455 /* ShanghaiTime */},
		},
		{
			"goerli",
			params.GoerliChainConfig,
			[]uint64{1_561_651, 4_460_644, 5_062_605},
			[]uint64{1678832736 /* ShanghaiTime */},
		},
		{
			"sepolia",
			params.SepoliaChainConfig,
			[]uint64{1735371 /* Merge NetSplit */},
			[]uint64{1677557088 /* ShanghaiTime */},
		},
		{
			"mordor",
			params.MordorChainConfig,
			[]uint64{301_243, 999_983, 2_520_000, 3_985_893, 5_520_000, 9_957_000},
			[]uint64{},
		},
		{
			"mintme",
			params.MintMeChainConfig,
			[]uint64{252_500},
			[]uint64{},
		},
	}
	sliceContains := func(sl []uint64, u uint64) bool {
		for _, s := range sl {
			if s == u {
				return true
			}
		}
		return false
	}
	slicesEquivalent := func(sl1, sl2 []uint64) bool {
		if len(sl1) != len(sl2) {
			return false
		}
		for _, s := range sl1 {
			if !sliceContains(sl2, s) {
				return false
			}
		}
		return true
	}
	for _, c := range cases {
		blockForks, timeForks := gatherForks(c.config)
		if !slicesEquivalent(blockForks, c.blockForks) {
			t.Errorf("config=%s block forks mismatch: have %v, want %v", c.name, blockForks, c.blockForks)
		}
		if !slicesEquivalent(timeForks, c.timeForks) {
			t.Errorf("config=%s time forks mismatch: have %v, want %v", c.name, timeForks, c.timeForks)
		}
	}
}

// TestGenerateSpecificationCases generates markdown formatted specification
// for network forkid values.
func TestGenerateSpecificationCases(t *testing.T) {
	if os.Getenv("COREGETH_GENERATE_FORKID_TEST_CASES") == "" {
		t.Skip()
	}
	type testCaseJSON struct {
		ChainConfig *coregeth.CoreGethChainConfig `json:"geth_chain_config"`
		GenesisHash common.Hash                   `json:"genesis_hash"`
		Head        uint64                        `json:"head"`
		ForkHash    common.Hash                   `json:"fork_hash"`
		ForkNext    uint64                        `json:"fork_next"`
		ForkIDRLP   common.Hash                   `json:"fork_id_rlp"`
	}

	generatedCases := []*testCaseJSON{}

	tests := []struct {
		name        string
		config      ctypes.ChainConfigurator
		genesisHash common.Hash
	}{
		{"Ethereum Classic Mainnet (ETC)",
			params.ClassicChainConfig,
			params.MainnetGenesisHash,
		},
		{
			"Mordor",
			params.MordorChainConfig,
			params.MordorGenesisHash,
		},
		{
			"Morden",
			&coregeth.CoreGethChainConfig{
				Ethash:            &ctypes.EthashConfig{},
				EIP2FBlock:        big.NewInt(494000),
				EIP150Block:       big.NewInt(1783000),
				EIP155Block:       big.NewInt(1915000),
				ECIP1017FBlock:    big.NewInt(2000000),
				ECIP1017EraRounds: big.NewInt(2000000),
				DisposalBlock:     big.NewInt(2300000),
				EIP198FBlock:      big.NewInt(4729274), // Atlantis
				EIP1052FBlock:     big.NewInt(5000381), // Agharta
			},
			common.HexToHash("0cd786a2425d16f152c658316c423e6ce1181e15c3295826d7c9904cba9ce303"),
		},
		{
			"MintMe",
			params.MintMeChainConfig,
			params.MintMeGenesisHash,
		},
	}
	for _, tt := range tests {
		cs := []uint64{0}
		blockForks, _ := gatherForks(tt.config)
		for _, f := range blockForks {
			cs = append(cs, f-1, f, f+1)
		}
		fmt.Printf("##### %s\n", tt.name)
		fmt.Println()
		fmt.Printf("- Genesis Hash: `0x%x`\n", tt.genesisHash)
		forks, _ := gatherForks(tt.config)
		forksS := []string{}
		for _, fi := range forks {
			forksS = append(forksS, strconv.Itoa(int(fi)))
		}
		fmt.Printf("- BlockForks: `%s`\n", strings.Join(forksS, "`,`"))
		fmt.Println()
		fmt.Println("| Head Block Number | `FORK_HASH` | `FORK_NEXT` | RLP Encoded (Hex) |")
		fmt.Println("| --- | --- | --- | --- |")
		for _, c := range cs {
			id := NewID(tt.config, tt.genesisHash, c, 0)
			isCanonical := false
			for _, fi := range forks {
				if c == fi {
					isCanonical = true
				}
			}
			r, _ := rlp.EncodeToBytes(id)
			if isCanonical {
				fmt.Printf("| __head=%d__ | FORK_HASH=%x | FORK_NEXT=%d | %x |\n", c, id.Hash, id.Next, r)
			} else {
				fmt.Printf("| head=%d | FORK_HASH=%x | FORK_NEXT=%d | %x |\n", c, id.Hash, id.Next, r)
			}

			gethConfig := &coregeth.CoreGethChainConfig{}
			err := confp.Crush(gethConfig, tt.config, true)
			if err != nil {
				t.Fatal(err)
			}
			generatedCases = append(generatedCases, &testCaseJSON{
				ChainConfig: gethConfig,
				GenesisHash: tt.genesisHash,
				Head:        c,
				ForkHash:    common.BytesToHash(id.Hash[:]),
				ForkNext:    id.Next,
				ForkIDRLP:   common.BytesToHash(r),
			})
		}
		fmt.Println()
		fmt.Println()
		t.Log("generated cases", generatedCases)
	}
}