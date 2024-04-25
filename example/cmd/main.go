package main

import (
	"context"
	"math/big"
	"os"
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
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	geth_core "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	chainId       = big.NewInt(1337)
	privateKeyHex = "2fc96e918d52d60d78657d7c8b021207ae5cd7d20a311363b16d6bc08f6efd78"
	pcAddress     = common.HexToAddress("0x1234")
)

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

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

	specs := arch.ArchSchemas{Actions: archmod.ActionSpecs, Tables: archmod.TableSpecs}
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

	nonce, err := sim.PendingNonceAt(context.Background(), from)
	if err != nil {
		panic(err)
	}

	var (
		kv          = kvstore.NewMemoryKeyValueStore()
		blockTime   = 1 * time.Second
		blockNumber = uint64(0)
	)

	var (
		actionChan              = make(chan []arch.Action, 8)
		actionBatchChan         = make(chan arch.ActionBatch, 8)
		actionBatchChanDampened = make(chan arch.ActionBatch, 1)
		txHashChan              = make(chan common.Hash, 1)
		txUpdateChanW           = make(chan *rpc.ActionTxUpdate, 1)
	)

	sender := rpc.NewActionSender(sim, specs.Actions, nil, gameAddress, from, nonce, signerFn)
	_, cancel := sender.StartSendingActions(actionChan, txUpdateChanW)
	defer cancel()

	sub := rpc.SubscribeActionBatches(sim, specs.Actions, coreAddress, 0, actionBatchChan, txHashChan)
	defer sub.Unsubscribe()
	rpc.DampenLatency(actionBatchChan, actionBatchChanDampened, blockTime, 100*time.Millisecond)

	go func() {
		for txHash := range txHashChan {
			txUpdateChanW <- &rpc.ActionTxUpdate{
				TxHash: txHash,
				Status: rpc.ActionTxStatus_Included,
			}
		}
	}()

	txUpdateChanR := utils.ProbeChannel(txUpdateChanW, func(txUpdate *rpc.ActionTxUpdate) {
		log.Info("Transaction "+txUpdate.Status.String(), "nonce", txUpdate.Nonce, "txHash", txUpdate.TxHash.Hex())
	})

	hinter := rpc.NewTxHinter(sim, txUpdateChanR)
	hinter.Start(blockTime / 2)

	sim.Start(blockTime, gameAddress)
	defer sim.Stop()

	c := client.NewClient(kv, actionBatchChanDampened, actionChan, blockTime, blockNumber, hinter)
	w, h := c.Layout(-1, -1)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Archetype Example")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		panic(err)
	}
}
