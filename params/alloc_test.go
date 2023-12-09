package params

import (
	"fmt"

	"github.com/yuriy0803/core-geth1/common"
	"github.com/yuriy0803/core-geth1/params/types/genesisT"
)

func ExampleMainnetAllocData() {
	// Test that the mainnet alloc is parsable.
	alloc := MainnetAllocData
	ga := genesisT.DecodePreAlloc(alloc)

	fmt.Println(ga[common.Address{0x3}])
	fmt.Println(ga[common.HexToAddress("0x3000000000000000000000000000000000000003")])
	// Output:
	// {[] map[] <nil> 0 []}
	// {[] map[] <nil> 0 []}
}
