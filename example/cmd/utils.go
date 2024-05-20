package cmd

import (
	"context"
	"errors"
	"math/big"
	"time"

	game_contract "github.com/concrete-eth/archetype/example/gogen/abigen/game"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func WaitForTx(ethcli *ethclient.Client, tx *types.Transaction) error {
	timeout := time.After(6 * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("timeout")
		default:
		}
		_, pending, err := ethcli.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			return err
		} else if pending {
			time.Sleep(1 * time.Second)
			continue
		} else {
			return nil
		}
	}
}

func SetNonce(ethcli *ethclient.Client, auth *bind.TransactOpts) error {
	nonce, err := ethcli.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return err
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	return nil
}

func DeployGame(ethcli *ethclient.Client, auth *bind.TransactOpts, pcAddr common.Address) (gameAddr common.Address, coreAddr common.Address, err error) {
	var tx *types.Transaction
	var gameContract *game_contract.Contract
	SetNonce(ethcli, auth)
	gameAddr, tx, gameContract, err = game_contract.DeployContract(auth, ethcli)
	if err != nil {
		return
	}
	if err := WaitForTx(ethcli, tx); err != nil {
		return gameAddr, coreAddr, err
	}

	SetNonce(ethcli, auth)
	tx, err = gameContract.Initialize(auth, pcAddr)
	if err != nil {
		return
	}
	if err := WaitForTx(ethcli, tx); err != nil {
		return gameAddr, coreAddr, err
	}

	coreAddr, err = gameContract.Proxy(nil)
	if err != nil {
		return
	}

	return gameAddr, coreAddr, nil
}

func NewEthClient(rpcUrl string) (*ethclient.Client, *big.Int, error) {
	ethcli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		panic(err)
	}
	chainId, err := ethcli.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	return ethcli, chainId, nil
}
