package main

import (
	"fmt"
	"sync"

	"github.com/axiomesh/data-producer/internal/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var balanceCMD = &cli.Command{
	Name:  "balance",
	Usage: "Accounts' Balance tools",
	Subcommands: []*cli.Command{
		{
			Name:  "init",
			Usage: "Init Accounts' Balance",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "amount",
					Usage: "Specify account's amount",
				},
			},
			Action: InitAccountsBalance,
		},
	},
}

func InitAccountsBalance(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	accounts, err := IsInitAccounts()
	if err != nil {
		return err
	}

	quantity := len(accounts)
	var parallel, cnt int
	if quantity <= DefaultParallel {
		parallel = 1
		cnt = quantity
	} else {
		parallel = DefaultParallel
		cnt = quantity / parallel
		if quantity%parallel != 0 {
			parallel++
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			var end int
			if idx == parallel-1 {
				end = quantity
			} else {
				end = (idx + 1) * cnt
			}

			for j := idx * cnt; j < end; j++ {
				key, err := crypto.HexToECDSA(accounts[j])
				if err != nil {
					fmt.Println(err)
					return
				}
				err = utils.TransferFromAdmin(key)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	return nil
}
