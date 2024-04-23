package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/client"
	"github.com/concrete-eth/archetype/example/core"
	game_contract "github.com/concrete-eth/archetype/example/gogen/abigen/game"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/concrete-eth/archetype/sim"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	geth_core "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	chainId       = big.NewInt(1337)
	privateKeyHex = "2fc96e918d52d60d78657d7c8b021207ae5cd7d20a311363b16d6bc08f6efd78"
	pcAddress     = common.HexToAddress("0x1234")
)

func main() {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		panic(err)
	}
	from := auth.From
	signerFn := auth.Signer

	specs := arch.ArchSpecs{Actions: archmod.ActionSpecs, Tables: archmod.TableSpecs}
	pc := precompile.NewCorePrecompile(specs, &core.Core{})

	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, pcAddress, pc)

	alloc := geth_core.GenesisAlloc{from: {Balance: new(big.Int).SetUint64(1e18)}}
	sim := sim.NewTickingSimulatedBackend(alloc, 1e8, registry)

	gameAddress, tx, gameContract, err := game_contract.DeployContract(auth, sim)
	if err != nil {
		panic(err)
	}
	sim.Commit()

	if receipt, err := sim.TransactionReceipt(context.Background(), tx.Hash()); err != nil {
		panic(err)
	} else if receipt.Status != 1 {
		panic("contract deployment failed")
	}

	tx, err = gameContract.Initialize(auth, pcAddress)
	if err != nil {
		panic(err)
	}
	sim.Commit()

	if receipt, err := sim.TransactionReceipt(context.Background(), tx.Hash()); err != nil {
		panic(err)
	} else if receipt.Status != 1 {
		panic("contract deployment failed")
	}

	coreAddress, err := gameContract.Proxy(nil)
	if err != nil {
		panic(err)
	}
	sim.Commit()

	var (
		kv               = kvstore.NewMemoryKeyValueStore()
		actionBatchChan  = make(chan arch.ActionBatch, 1)
		txHashChan       = make(chan common.Hash, 1) // TODO: better naming
		actionOutChan    = make(chan []arch.Action, 1)
		txUpdateOutChan0 = make(chan *rpc.ActionTxUpdate, 1)
		txUpdateOutChan1 = make(chan *rpc.ActionTxUpdate, 1)
		blockTime        = 1 * time.Second
		blockNumber      = uint64(0)
	)

	sub := rpc.SubscribeActionBatches(sim, specs.Actions, coreAddress, 0, actionBatchChan, txHashChan)
	defer sub.Unsubscribe()

	nonce, err := sim.PendingNonceAt(context.Background(), from)
	if err != nil {
		panic(err)
	}

	sender := rpc.NewActionSender(sim, specs.Actions, nil, gameAddress, from, nonce, signerFn)
	_, cancel := sender.StartSendingActions(actionOutChan, txUpdateOutChan0)
	defer cancel()

	go func() {
		for txUpdate := range txUpdateOutChan0 {
			fmt.Println("tx update:", txUpdate.Nonce, txUpdate.TxHash.Hex(), txUpdate.Status, txUpdate.Err)
			txUpdateOutChan1 <- txUpdate
		}
	}()

	sim.Start(blockTime, gameAddress)
	defer sim.Stop()

	hinter := rpc.NewTxHinter(sim, txUpdateOutChan1)
	hinter.Start(blockTime / 2)

	go func() {
		for txHash := range txHashChan {
			txUpdateOutChan1 <- &rpc.ActionTxUpdate{
				TxHash: txHash,
				Status: rpc.ActionTxStatus_Included,
			}
		}
	}()

	c := client.NewClient(kv, actionBatchChan, actionOutChan, blockTime, blockNumber, hinter)
	w, h := c.Layout(0, 0)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Archetype Example")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		panic(err)
	}
}
