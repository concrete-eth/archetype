package deploy

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/concrete-eth/archetype/rpc"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewEthClient(rpcUrl string) (ethcli *ethclient.Client, chainId *big.Int, err error) {
	if ethcli, err = ethclient.Dial(rpcUrl); err != nil {
		return
	} else if chainId, err = ethcli.ChainID(context.Background()); err != nil {
		return
	}
	return
}

func SetNonce(auth *bind.TransactOpts, ethcli ethereum.PendingStateReader) error {
	if nonce, err := ethcli.PendingNonceAt(context.Background(), auth.From); err != nil {
		return err
	} else {
		auth.Nonce = new(big.Int).SetUint64(nonce)
	}
	return nil
}

func WaitForTx(ethcli ethereum.TransactionReader, tx *types.Transaction) error {
	timeout := time.After(1 * time.Second)
	for {
		_, pending, err := ethcli.TransactionByHash(context.Background(), tx.Hash())
		if err == ethereum.NotFound {
			select {
			case <-timeout:
				return errors.New("timeout")
			default:
			}
		} else if err != nil {
			return err
		} else if !pending {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

type GameContractDeployer = func(auth *bind.TransactOpts, ethcli bind.ContractBackend) (common.Address, *types.Transaction, InitializableProxyAdmin, error)

type InitializableProxyAdmin interface {
	Proxy(opts *bind.CallOpts) (common.Address, error)
	Initialize(auth *bind.TransactOpts, logic common.Address) (*types.Transaction, error)
}

func DeployGame(auth *bind.TransactOpts, ethcli rpc.EthCli, deployer GameContractDeployer, logic common.Address, commit bool) (gameAddr common.Address, coreAddr common.Address, err error) {
	var tx *types.Transaction
	var proxyAdmin InitializableProxyAdmin

	SetNonce(auth, ethcli)
	if gameAddr, tx, proxyAdmin, err = deployer(auth, ethcli); err != nil {
		return
	} else {
		if commit {
			ethcli.(interface{ Commit() }).Commit()
		}
		if err = WaitForTx(ethcli, tx); err != nil {
			return
		}
	}

	SetNonce(auth, ethcli)
	if tx, err = proxyAdmin.Initialize(auth, logic); err != nil {
		return
	} else {
		if commit {
			ethcli.(interface{ Commit() }).Commit()
		}
		if err = WaitForTx(ethcli, tx); err != nil {
			return
		}
	}

	if coreAddr, err = proxyAdmin.Proxy(nil); err != nil {
		return
	}

	return gameAddr, coreAddr, nil
}
