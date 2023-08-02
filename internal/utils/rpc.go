package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const AdminPrivateKey = "b6477143e17f889263044f6cf463dc37177ac4526c4c39a7a344198457024a2f" // axiom admin // axiom json rpc
const ContractABI = "[{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"num\",\"type\":\"uint64\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
const ContractBIN = "608060405234801561001057600080fd5b50610186806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80631d9a3bdd1461003b5780632e64cec114610057575b600080fd5b610055600480360381019061005091906100d2565b610075565b005b61005f6100a0565b60405161006c919061010a565b60405180910390f35b806000806101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555050565b60008060009054906101000a900467ffffffffffffffff16905090565b6000813590506100cc81610139565b92915050565b6000602082840312156100e457600080fd5b60006100f2848285016100bd565b91505092915050565b61010481610125565b82525050565b600060208201905061011f60008301846100fb565b92915050565b600067ffffffffffffffff82169050919050565b61014281610125565b811461014d57600080fd5b5056fea26469706673582212204691849347a2f1bef4241dc7ceacb8f17c8556dad30e2a1f78e8450c908986a764736f6c63430008040033"

var client *ethclient.Client
var nonce uint64

func InitClient(url string) error {
	// init client
	rpc, err := ethclient.Dial(url)
	if err != nil {
		return err
	}
	client = rpc
	// init nonce
	adminKey, err := crypto.HexToECDSA(AdminPrivateKey)
	if err != nil {
		return err
	}
	from := crypto.PubkeyToAddress(adminKey.PublicKey)
	nonce, err = client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return err
	}
	nonce--
	return nil
}

func TransferFromAdmin(key *ecdsa.PrivateKey) error {
	adminKey, err := crypto.HexToECDSA(AdminPrivateKey)
	if err != nil {
		return err
	}
	to := crypto.PubkeyToAddress(key.PublicKey)
	// 1 BXH
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    atomic.AddUint64(&nonce, 1),
		To:       &to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     []byte{},
	})
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), adminKey)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signTx)
	if err != nil {
		return err
	}

	hash := signTx.Hash()
	err = retry.Retry(func(attempt uint) error {
		receipt, err := client.TransactionReceipt(context.Background(), hash)
		if err != nil {
			return err
		}
		if receipt.Status != uint64(1) {
			return fmt.Errorf("transfer from admin error")
		}
		return nil
	}, strategy.Limit(3), strategy.Delay(time.Millisecond*500))
	if err != nil {
		return err
	}
	return nil
}

func GetBlockHighMax() (int, error) {
	number, err := client.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}
	return int(number), nil
}

func GetBlockRandomHash() (string, error) {
	number, err := client.BlockNumber(context.Background())
	if err != nil {
		return "", err
	}
	idx := rand.Intn(int(number)) + 1
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(idx)))
	if err != nil {
		return "", err
	}
	return block.Hash().String(), nil
}

func GetTxRandomHash() (string, error) {
	number, err := client.BlockNumber(context.Background())
	if err != nil {
		return "", err
	}
	idx := rand.Intn(int(number)) + 1
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(idx)))
	if err != nil {
		return "", err
	}
	if block.Transactions().Len() == 0 {
		return GetTxRandomHash()
	}
	return block.Transactions()[0].Hash().String(), nil
}

func DeployContract() (string, error) {
	bytecode := common.Hex2Bytes(ContractBIN)
	adminKey, err := crypto.HexToECDSA(AdminPrivateKey)
	if err != nil {
		return "", err
	}
	gasLimit := uint64(210000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    atomic.AddUint64(&nonce, 1),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     bytecode,
	})
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), adminKey)
	if err != nil {
		return "", err
	}
	err = client.SendTransaction(context.Background(), signTx)
	if err != nil {
		return "", err
	}

	hash := signTx.Hash()
	var address string
	err = retry.Retry(func(attempt uint) error {
		receipt, err := client.TransactionReceipt(context.Background(), hash)
		if err != nil {
			return err
		}
		if receipt.Status != uint64(1) {
			return fmt.Errorf("deploy contract error")
		}
		address = receipt.ContractAddress.String()
		return nil
	}, strategy.Limit(3), strategy.Delay(time.Millisecond*500))
	if err != nil {
		return "", err
	}
	return address, nil
}

func Store(address string, value uint64) error {
	loadABI, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return err
	}
	pack, err := loadABI.Pack("store", value)
	if err != nil {
		return err
	}
	adminKey, err := crypto.HexToECDSA(AdminPrivateKey)
	if err != nil {
		return err
	}
	gasLimit := uint64(210000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}
	toAddress := common.HexToAddress(address)

	tx := types.NewTx(&types.LegacyTx{
		To:       &toAddress,
		Nonce:    atomic.AddUint64(&nonce, 1),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     pack,
	})
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), adminKey)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signTx)
	if err != nil {
		return err
	}

	hash := signTx.Hash()
	err = retry.Retry(func(attempt uint) error {
		receipt, err := client.TransactionReceipt(context.Background(), hash)
		if err != nil {
			return err
		}
		if receipt.Status != uint64(1) {
			return fmt.Errorf("invoke contract store error")
		}
		return nil
	}, strategy.Limit(3), strategy.Delay(time.Millisecond*500))
	if err != nil {
		return err
	}
	return nil
}

func GenEstimateGasTx(contractAddr, privateKey string, value uint64) (*types.Transaction, error) {
	loadABI, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	pack, err := loadABI.Pack("store", value)
	if err != nil {
		return nil, err
	}
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	toAddress := common.HexToAddress(contractAddr)

	tx := types.NewTx(&types.LegacyTx{
		To:    &toAddress,
		Nonce: nonce,
		Data:  pack,
	})
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		return nil, err
	}
	return signTx, nil
}

func GenRawTx(contractAddr, privateKey string, value uint64) (*types.Transaction, error) {
	loadABI, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	pack, err := loadABI.Pack("store", value)
	if err != nil {
		return nil, err
	}
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	gasLimit := uint64(210000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	toAddress := common.HexToAddress(contractAddr)

	tx := types.NewTx(&types.LegacyTx{
		To:       &toAddress,
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     pack,
	})
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		return nil, err
	}
	return signTx, nil
}

func GenRetrieveMsg(address, from string) (*ethereum.CallMsg, error) {
	loadABI, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	pack, err := loadABI.Pack("retrieve")
	if err != nil {
		return nil, err
	}
	toAddress := common.HexToAddress(address)

	msg := &ethereum.CallMsg{
		From: common.HexToAddress(from),
		To:   &toAddress,
		Data: pack,
	}
	return msg, nil
}
