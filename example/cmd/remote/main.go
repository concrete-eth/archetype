package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/client"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	rpcUrl   = "ws://localhost:9546"
	gameAddr = common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	coreAddr = common.HexToAddress("0xa16E02E87b7454126E5E10d957A927A7F5B5d2be")
)

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))

	// Dial rpc
	ethcli, chainId, err := rpc.NewEthClient(rpcUrl)
	if err != nil {
		panic(err)
	}

	// Load transaction signing credentials from the process environment.
	privateKeyHex := strings.TrimPrefix(strings.TrimSpace(os.Getenv("ARCHETYPE_DEPLOY_PRIVATE_KEY")), "0x")
	if privateKeyHex == "" {
		panic("ARCHETYPE_DEPLOY_PRIVATE_KEY must be set")
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		panic(err)
	}
	rpc.SetNonce(auth, ethcli)

	// Create schemas from codegen
	schemas := arch.ArchSchemas{Actions: archmod.ActionSchemas, Tables: archmod.TableSchemas}

	// Create chain IO
	var (
		blockTime           = 2 * time.Second
		startingBlockNumber = uint64(0)
		dampenDelay         = 100 * time.Millisecond
	)
	io := rpc.NewIO(ethcli, blockTime, schemas, auth, gameAddr, coreAddr, startingBlockNumber, dampenDelay)
	io.SetTxUpdateHook(func(txUpdate *rpc.ActionTxUpdate) {
		log.Info("Transaction "+txUpdate.Status.String(), "nonce", txUpdate.Nonce, "txHash", txUpdate.TxHash.Hex())
	})
	defer io.Stop()

	// Create and start client
	kv := kvstore.NewMemoryKeyValueStore()
	c := client.NewClient(kv, io)
	if bn, err := ethcli.BlockNumber(context.Background()); err != nil {
		panic(err)
	} else {
		c.SyncUntil(bn)
	}
	w, h := c.Layout(-1, -1)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Archetype Example")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		panic(err)
	}
}
