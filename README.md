# data-producer

A stress test tool for [@axiom](https://github.com/axiomesh/axiom)'s Ethereum JSON RPC🚀

## Usage

Go install project

```shell
make install
```

Command Interface

```
NAME:
   data - produce stress test data

USAGE:
   data [global options] command [command options] [arguments...]

COMMANDS:
   init      Init Default Accounts
   balance   Accounts' Balance tools
   generate  Generate stress testing testdata
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

```

## JSON RPC testdata mock support

* `eth_getBalance`
* `eth_getBlockByNumber`
* `eth_getBlockByNumber`
* `eth_getBlockByHash`
* `eth_getCode`
* `eth_getStorageAt`
* `eth_call`
* `eth_estimateGas`
* `eth_getBlockTransactionCountByNumber`
* `eth_getBlockTransactionCountByHash`
* `eth_getTransactionByBlockNumberAndIndex`
* `eth_getTransactionByBlockHashAndIndex`
* `eth_getTransactionCount`
* `eth_getTransactionByHash`
* `eth_getTransactionReceipt`
* `eth_sendRawTransaction`