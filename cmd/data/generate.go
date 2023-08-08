package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/axiomesh/data-producer/internal/model"
	"github.com/axiomesh/data-producer/internal/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var generateCMD = &cli.Command{
	Name:  "generate",
	Usage: "Generate stress testing testdata",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "quantity",
			Usage: "Specify testdata quantity, should less or equal account number, Max(2000000)",
			Value: DefaultQuantity,
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "eth_getBalance",
			Usage:  "Generate eth_getBalance's testdata",
			Action: generateEthGetBalance,
		},
		{
			Name:   "eth_getBlockByNumber",
			Usage:  "Generate eth_getBlockByNumber's testdata",
			Action: generateEthGetBlockByNumber,
		},
		{
			Name:   "eth_getBlockByHash",
			Usage:  "Generate eth_getBlockByHash's testdata",
			Action: generateEthGetBlockByHash,
		},
		{
			Name:   "eth_getCode",
			Usage:  "Generate eth_getCode's testdata",
			Action: generateEthGetCode,
		},
		{
			Name:   "eth_getStorageAt",
			Usage:  "Generate eth_getStorageAt's testdata",
			Action: generateEthGetStorageAt,
		},
		{
			Name:   "eth_call",
			Usage:  "Generate eth_call's testdata",
			Action: generateEthCall,
		},
		{
			Name:   "eth_estimateGas",
			Usage:  "Generate eth_estimateGas's testdata",
			Action: generateEthEstimateGas,
		},
		{
			Name:   "eth_getBlockTransactionCountByNumber",
			Usage:  "Generate eth_getBlockTransactionCountByNumber's testdata",
			Action: generateEthGetBlockTransactionCountByNumber,
		},
		{
			Name:   "eth_getBlockTransactionCountByHash",
			Usage:  "Generate eth_getBlockTransactionCountByHash's testdata",
			Action: generateEthGetBlockTransactionCountByHash,
		},
		{
			Name:   "eth_getTransactionByBlockNumberAndIndex",
			Usage:  "Generate eth_getTransactionByBlockNumberAndIndex's testdata",
			Action: generateEthGetTransactionByBlockNumberAndIndex,
		},
		{
			Name:   "eth_getTransactionByBlockHashAndIndex",
			Usage:  "Generate eth_getTransactionByBlockHashAndIndex's testdata",
			Action: generateEthGetTransactionByBlockHashAndIndex,
		},
		{
			Name:   "eth_getTransactionCount",
			Usage:  "Generate eth_getTransactionCount's testdata",
			Action: generateEthGetTransactionCount,
		},
		{
			Name:   "eth_getTransactionByHash",
			Usage:  "Generate eth_getTransactionByHash's testdata",
			Action: generateEthGetTransactionByHash,
		},
		{
			Name:   "eth_getTransactionReceipt",
			Usage:  "Generate eth_getTransactionReceipt's testdata",
			Action: generateEthGetTransactionReceipt,
		},
		{
			Name:   "eth_sendRawTransaction",
			Usage:  "Generate eth_sendRawTransaction's testdata",
			Action: generateEthSendRawTransaction,
		},
	},
}

func generate(ctx *cli.Context, fn func(account string) (*model.EthReq, error)) error {
	quantity := ctx.Int("quantity")
	if quantity > DefaultQuantity {
		return fmt.Errorf("quantity is large than max quantity(%d)", DefaultQuantity)
	}
	accounts, err := IsInitAccounts()
	if err != nil {
		return err
	}
	if len(accounts) < quantity {
		return fmt.Errorf("quantity is large than account number(%d)", len(accounts))
	}
	var reqs []*model.EthReq
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
	lock := sync.Mutex{}
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
				req, err := fn(accounts[j])
				if err != nil {
					fmt.Println(err)
					return
				}
				lock.Lock()
				reqs = append(reqs, req)
				lock.Unlock()
			}
		}(i)
	}
	wg.Wait()

	filename := fmt.Sprintf("%s-%d-%s.csv", ctx.Command.Name, len(reqs), time.Now().Format("200601021504"))
	err = model.CreateAndWriteReqs(filename, reqs)
	if err != nil {
		return err
	}

	return nil
}

func generateEthGetBalance(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	fn := func(account string) (*model.EthReq, error) {
		key, err := crypto.HexToECDSA(account)
		if err != nil {
			return nil, err
		}
		address := crypto.PubkeyToAddress(key.PublicKey).String()
		req := model.NewGetBalanceReq(address)
		if err != nil {
			return nil, err
		}
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetBlockByNumber(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	highMax, err := utils.GetBlockHighMax()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		number := rand.Intn(highMax) + 1
		req := model.NewGetBlockByNumberReq(number)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetBlockByHash(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	fn := func(account string) (*model.EthReq, error) {
		hash, err := utils.GetBlockRandomHash()
		if err != nil {
			return nil, err
		}
		req := model.NewGetBlockByHashReq(hash)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetCode(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	contractAddr, err := utils.DeployContract()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewGetCodeReq(contractAddr)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetStorageAt(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	fn := func(account string) (*model.EthReq, error) {
		key, err := crypto.HexToECDSA(account)
		if err != nil {
			return nil, err
		}
		address := crypto.PubkeyToAddress(key.PublicKey).String()
		req := model.NewGetStorageAtReq(address)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthCall(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	// deploy contract
	contractAddr, err := utils.DeployContract()
	if err != nil {
		return err
	}
	// set value
	err = utils.Store(contractAddr, 1)
	if err != nil {
		return err
	}
	// get value
	fn := func(account string) (*model.EthReq, error) {
		key, err := crypto.HexToECDSA(account)
		if err != nil {
			return nil, err
		}
		address := crypto.PubkeyToAddress(key.PublicKey).String()
		msg, err := utils.GenRetrieveMsg(contractAddr, address)
		if err != nil {
			return nil, err
		}
		req := model.NewCallReq(msg)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthEstimateGas(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	// deploy contract
	contractAddr, err := utils.DeployContract()
	if err != nil {
		return err
	}
	// generate msg
	tx, err := utils.GenEstimateGasTx(contractAddr, utils.AdminPrivateKey, 1)
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewEstimateGasReq(tx)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetBlockTransactionCountByNumber(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	max, err := utils.GetBlockHighMax()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		number := rand.Intn(max) + 1
		req := model.NewGetBlockTransactionCountByNumberReq(number)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetBlockTransactionCountByHash(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	hash, err := utils.GetBlockRandomHash()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewGetBlockTransactionCountByHashReq(hash)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetTransactionByBlockNumberAndIndex(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	max, err := utils.GetBlockHighMax()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		number := rand.Intn(max) + 1
		req := model.NewGetTransactionByBlockNumberAndIndexReq(number, 0)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetTransactionByBlockHashAndIndex(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	hash, err := utils.GetBlockRandomHash()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewGetTransactionByBlockHashAndIndexReq(hash, 0)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetTransactionCount(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	fn := func(account string) (*model.EthReq, error) {
		key, err := crypto.HexToECDSA(account)
		if err != nil {
			return nil, err
		}
		address := crypto.PubkeyToAddress(key.PublicKey).String()
		req := model.NewGetTransactionCountReq(address)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetTransactionByHash(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	hash, err := utils.GetTxRandomHash()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewGetTransactionByHashReq(hash)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthGetTransactionReceipt(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	hash, err := utils.GetTxRandomHash()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		req := model.NewGetTransactionReceiptReq(hash)
		return req, nil
	}
	return generate(ctx, fn)
}

func generateEthSendRawTransaction(ctx *cli.Context) error {
	// init rpc client
	url := ctx.String("url")
	err := utils.InitClient(url)
	if err != nil {
		return err
	}

	// deploy contract
	contractAddr, err := utils.DeployContract()
	if err != nil {
		return err
	}
	fn := func(account string) (*model.EthReq, error) {
		// generate tx
		tx, err := utils.GenRawTx(contractAddr, account, 1)
		if err != nil {
			return nil, err
		}
		data, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		req := model.NewSendRawTransactionReq(hexutil.Encode(data))
		return req, nil
	}
	return generate(ctx, fn)
}
