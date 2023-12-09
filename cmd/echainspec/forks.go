package main

import (
	"fmt"

	"github.com/yuriy0803/core-geth1/params/confp"
	"gopkg.in/urfave/cli.v1"
)

var forksCommand = cli.Command{
	Name:   "forks",
	Usage:  "List unique and non-zero fork numbers",
	Action: forks,
}

func forks(ctx *cli.Context) error {
	for _, f := range confp.BlockForks(globalChainspecValue) {
		fmt.Println(f)
	}
	for _, f := range confp.TimeForks(globalChainspecValue) {
		fmt.Printf("@%d\n", f)
	}

	return nil
}
