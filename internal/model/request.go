package model

import (
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gocarina/gocsv"
)

type Params []interface{}

func (params *Params) MarshalCSV() (string, error) {
	bytes, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

type EthReq struct {
	ID      int    `json:"id" csv:"id"`
	JsonRpc string `json:"jsonrpc" csv:"jsonrpc"`
	Method  string `json:"method" csv:"method"`
	Params  Params `json:"params" csv:"params"`
}

func NewEthReq(method string, params ...interface{}) *EthReq {
	return &EthReq{
		ID:      1,
		JsonRpc: "2.0",
		Method:  method,
		Params:  params,
	}
}

func NewGetBalanceReq(address string) *EthReq {
	method := "eth_getBalance"
	return NewEthReq(method, address)
}

func NewGetBlockByNumberReq(number int) *EthReq {
	method := "eth_getBlockByNumber"
	hexNumber := hexutil.EncodeUint64(uint64(number))
	return NewEthReq(method, hexNumber, true)
}

func NewGetBlockByHashReq(hash string) *EthReq {
	method := "eth_getBlockByHash"
	return NewEthReq(method, hash)
}

func NewGetCodeReq(address string) *EthReq {
	method := "eth_getCode"
	return NewEthReq(method, address, "latest")
}

func NewGetStorageAtReq(address string) *EthReq {
	method := "eth_getStorageAt"
	return NewEthReq(method, address, "0x0", "latest")
}

func NewCallReq(msg *ethereum.CallMsg) *EthReq {
	toCallArgs := func(msg *ethereum.CallMsg) interface{} {
		arg := map[string]interface{}{
			"from": msg.From,
			"to":   msg.To,
		}
		if len(msg.Data) > 0 {
			arg["data"] = hexutil.Bytes(msg.Data)
		}
		if msg.Value != nil {
			arg["value"] = (*hexutil.Big)(msg.Value)
		}
		if msg.Gas != 0 {
			arg["gas"] = hexutil.Uint64(msg.Gas)
		}
		if msg.GasPrice != nil {
			arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
		}
		return arg
	}
	method := "eth_call"
	return NewEthReq(method, toCallArgs(msg))
}

func NewEstimateGasReq(tx *types.Transaction) *EthReq {
	method := "eth_estimateGas"
	return NewEthReq(method, tx, "latest")
}

func NewGetBlockTransactionCountByNumberReq(number int) *EthReq {
	method := "eth_getBlockTransactionCountByNumber"
	hexNumber := hexutil.EncodeUint64(uint64(number))
	return NewEthReq(method, hexNumber)
}

func NewGetBlockTransactionCountByHashReq(hash string) *EthReq {
	method := "eth_getBlockTransactionCountByHash"
	return NewEthReq(method, hash)
}

func NewGetTransactionByBlockNumberAndIndexReq(number, idx int) *EthReq {
	method := "eth_getTransactionByBlockNumberAndIndex"
	hexNumber := hexutil.EncodeUint64(uint64(number))
	hexIdx := hexutil.EncodeUint64(uint64(idx))
	return NewEthReq(method, hexNumber, hexIdx)
}

func NewGetTransactionByBlockHashAndIndexReq(hash string, idx int) *EthReq {
	method := "eth_getTransactionByBlockHashAndIndex"
	hexIdx := hexutil.EncodeUint64(uint64(idx))
	return NewEthReq(method, hash, hexIdx)
}

func NewGetTransactionCountReq(address string) *EthReq {
	method := "eth_getTransactionCount"
	return NewEthReq(method, address, "latest")
}

func NewGetTransactionByHashReq(hash string) *EthReq {
	method := "eth_getTransactionByHash"
	return NewEthReq(method, hash)
}

func NewGetTransactionReceiptReq(hash string) *EthReq {
	method := "eth_getTransactionReceipt"
	return NewEthReq(method, hash)
}

func NewSendRawTransactionReq(rawTx string) *EthReq {
	method := "eth_sendRawTransaction"
	return NewEthReq(method, rawTx)
}

func CreateAndWriteReqs(filename string, req []*EthReq) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	writer := gocsv.DefaultCSVWriter(file)
	err = gocsv.MarshalCSV(req, writer)
	if err != nil {
		return err
	}
	return nil
}
