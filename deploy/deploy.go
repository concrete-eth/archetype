package deploy

import (
	"context"
	"errors"
	"fmt"

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

func waitForSuccess(ethcli rpc.EthCli, tx *types.Transaction) error {
	if err := rpc.WaitForTx(ethcli, tx); err != nil {
		return err
	}
	receipt, err := ethcli.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return err
	}
	if receipt.Status != 1 {
		return errors.New("tx failed")
	}
	return nil
}

func DeployGame(auth *bind.TransactOpts, ethcli rpc.EthCli, deployer GameContractDeployer, logic common.Address, data []byte, commit bool) (gameAddr common.Address, coreAddr common.Address, err error) {
	var tx *types.Transaction
	var proxyAdmin InitializableProxyAdmin

	rpc.SetNonce(auth, ethcli)
	gameAddr, tx, proxyAdmin, err = deployer(auth, ethcli)
	if err != nil {
		return
	}
	if commit {
		ethcli.(interface{ Commit() }).Commit()
	}
	if err := waitForSuccess(ethcli, tx); err != nil {
		err = fmt.Errorf("deploy game contract failed: %w", err)
		return gameAddr, coreAddr, err
	}

	rpc.SetNonce(auth, ethcli)
	tx, err = proxyAdmin.Initialize(auth, logic, data)
	if err != nil {
		return
	}
	if commit {
		ethcli.(interface{ Commit() }).Commit()
	}
	if err := waitForSuccess(ethcli, tx); err != nil {
		err = fmt.Errorf("initialize game contract failed: %w", err)
		return gameAddr, coreAddr, err
	}

	return gameAddr, coreAddr, nil
}
