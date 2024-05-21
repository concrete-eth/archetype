package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type EthCli interface {
	bind.ContractBackend
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.PendingStateReader
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.TransactionSender
	BlockNumber(ctx context.Context) (uint64, error)
}
