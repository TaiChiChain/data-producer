package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/axiomesh/data-producer/internal/repo"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var initCMD = &cli.Command{
	Name:  "init",
	Usage: "Init Default Accounts",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "quantity",
			Usage: "Specify Initialize Accounts Quantity",
			Value: DefaultQuantity,
		},
	},
	Action: initAccounts,
}

func initAccounts(ctx *cli.Context) error {
	quantity := ctx.Int("quantity")
	parallel := DefaultParallel

	dir, err := repo.DirPath()
	if err != nil {
		return err
	}
	accountsPath, err := repo.AccountsPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(accountsPath); os.IsNotExist(err) {
		// create default dir
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}
		// create and write accounts
		err := createAndWriteAccounts(quantity, parallel, accountsPath)
		if err != nil {
			return err
		}
	} else {
		accounts, err := repo.LoadAccounts(accountsPath)
		if err != nil {
			return err
		}
		if len(accounts) != quantity {
			// create and write accounts
			err := createAndWriteAccounts(quantity, parallel, accountsPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createAndWriteAccounts(quantity, parallel int, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(file)
	defer func(buf *bufio.Writer) {
		err := buf.Flush()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(buf)
	var cnt int
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

	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			var end int
			if idx == parallel-1 {
				end = quantity
			} else {
				end = (idx + 1) * cnt
			}

			for j := idx * cnt; j < end; j++ {
				key, err := crypto.GenerateKey()
				if err != nil {
					fmt.Println(err)
					return
				}
				privateKeyHex := common.Bytes2Hex(crypto.FromECDSA(key))

				lock.Lock()
				_, err = io.WriteString(buf, privateKeyHex+"\n")
				lock.Unlock()
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func IsInitAccounts() ([]string, error) {
	accountsPath, err := repo.AccountsPath()
	if err != nil {
		return nil, err
	}
	accounts, err := repo.LoadAccounts(accountsPath)
	if err != nil {
		fmt.Println("accounts load failed")
		fmt.Println("please run init accounts first")
		return nil, err
	}
	return accounts, nil
}
