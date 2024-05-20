package main

import (
	"context"
	"errors"
	"math/big"
	"os"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/client"
	game_contract "github.com/concrete-eth/archetype/example/gogen/abigen/game"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/physics"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/concrete-eth/archetype/simulated"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	chainId       = big.NewInt(1337)
	privateKeyHex = "2fc96e918d52d60d78657d7c8b021207ae5cd7d20a311363b16d6bc08f6efd78"
	pcAddr        = common.HexToAddress("0x1234")
)

func newSimulatedBackend(schemas arch.ArchSchemas, devAddresses ...common.Address) *simulated.TickingSimulatedBackend {
	pc := precompile.NewCorePrecompile(schemas, func() arch.Core { return &physics.Core{} })
	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, pcAddr, pc)
	alloc := types.GenesisAlloc{}
	for _, addr := range devAddresses {
		alloc[addr] = types.Account{Balance: new(big.Int).SetUint64(1e18)}
	}
	return simulated.NewTickingSimulatedBackend(alloc, 1e8, registry)
}

func deployGame(ethcli *simulated.TickingSimulatedBackend, auth *bind.TransactOpts) (gameAddr common.Address, coreAddr common.Address, err error) {
	var tx *types.Transaction
	var gameContract *game_contract.Contract
	gameAddr, tx, gameContract, err = game_contract.DeployContract(auth, ethcli)
	if err != nil {
		return
	}
	ethcli.Commit()

	var receipt *types.Receipt

	receipt, err = ethcli.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return
	} else if receipt.Status != 1 {
		return gameAddr, coreAddr, errors.New("contract deployment failed")
	}

	tx, err = gameContract.Initialize(auth, pcAddr)
	if err != nil {
		return
	}
	ethcli.Commit()

	receipt, err = ethcli.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return
	} else if receipt.Status != 1 {
		return gameAddr, coreAddr, errors.New("contract initialization failed")
	}

	coreAddr, err = gameContract.Proxy(nil)
	if err != nil {
		return
	}
	// ethcli.Commit()

	return gameAddr, coreAddr, nil
}

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))

	// Load tx opts
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		panic(err)
	}

	// Create schemas from codegen
	schemas := arch.ArchSchemas{Actions: archmod.ActionSchemas, Tables: archmod.TableSchemas}

	// Create simulated backend with precompile
	ethcli := newSimulatedBackend(schemas, auth.From)

	// Deploy game
	gameAddr, coreAddr, err := deployGame(ethcli, auth)
	if err != nil {
		panic(err)
	}

	// Set nonce
	nonce, err := ethcli.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		panic(err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Create chain IO
	var (
		blockTime           = 1 * time.Second
		startingBlockNumber = uint64(0)
	)
	io := rpc.NewIO(ethcli, blockTime, schemas, auth, gameAddr, coreAddr, startingBlockNumber, 0)
	io.SetTxUpdateHook(func(txUpdate *rpc.ActionTxUpdate) {
		log.Info("Transaction "+txUpdate.Status.String(), "nonce", txUpdate.Nonce, "txHash", txUpdate.TxHash.Hex())
	})

	// Start simulated ticking
	ethcli.Start(blockTime, gameAddr)
	defer ethcli.Stop()

	// Create and start client
	kv := kvstore.NewMemoryKeyValueStore()
	c := client.NewClient(kv, io)
	w, h := c.Layout(-1, -1)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Archetype Example")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		panic(err)
	}
}
