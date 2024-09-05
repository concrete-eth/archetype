package deploy

import (
	"context"
	"errors"

	"github.com/concrete-eth/archetype/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type GameContractDeployer = func(auth *bind.TransactOpts, ethcli bind.ContractBackend) (common.Address, *types.Transaction, InitializableProxyAdmin, error)

type InitializableProxyAdmin interface {
	Proxy(opts *bind.CallOpts) (common.Address, error)
	Initialize(auth *bind.TransactOpts, logic common.Address, data []byte) (*types.Transaction, error)
}

func DeployGame(auth *bind.TransactOpts, ethcli rpc.EthCli, deployer GameContractDeployer, logic common.Address, data []byte, commit bool) (gameAddr common.Address, coreAddr common.Address, err error) {
	var tx *types.Transaction
	var proxyAdmin InitializableProxyAdmin

	rpc.SetNonce(auth, ethcli)
	if gameAddr, tx, proxyAdmin, err = deployer(auth, ethcli); err != nil {
		return
	} else {
		if commit {
			ethcli.(interface{ Commit() }).Commit()
		}
		if err := rpc.WaitForTx(ethcli, tx); err != nil {
			return gameAddr, coreAddr, err
		}
		receipt, err := ethcli.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return gameAddr, coreAddr, err
		}
		if receipt.Status != 1 {
			return gameAddr, coreAddr, errors.New("deploy tx failed")
		}
	}

	rpc.SetNonce(auth, ethcli)
	if tx, err = proxyAdmin.Initialize(auth, logic, data); err != nil {
		return
	} else {
		if commit {
			ethcli.(interface{ Commit() }).Commit()
		}
		if err := rpc.WaitForTx(ethcli, tx); err != nil {
			return gameAddr, coreAddr, err
		}
		receipt, err := ethcli.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return gameAddr, coreAddr, err
		}
		if receipt.Status != 1 {
			return gameAddr, coreAddr, errors.New("initialize tx failed")
		}
	}

	if coreAddr, err = proxyAdmin.Proxy(nil); err != nil {
		return
	}

	return gameAddr, coreAddr, nil
}
