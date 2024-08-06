package main

import (
	"os"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/deploy"
	"github.com/concrete-eth/archetype/example/client"
	game_contract "github.com/concrete-eth/archetype/example/gogen/abigen/game"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/physics"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	pcAddr = common.HexToAddress("0x1234")
)

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))

	// Create schemas from codegen
	schemas := arch.ArchSchemas{Actions: archmod.ActionSchemas, Tables: archmod.TableSchemas}

	// Create precompile
	pc := precompile.NewCorePrecompile(schemas, func() arch.Core { return &physics.Core{} })
	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, pcAddr, pc)

	// Create local simulated io
	io, err := deploy.NewLocalIO(registry, schemas, func(auth *bind.TransactOpts, ethcli bind.ContractBackend) (common.Address, *types.Transaction, deploy.InitializableProxyAdmin, error) {
		return game_contract.DeployContract(auth, ethcli)
	}, pcAddr, nil, 1*time.Second)
	if err != nil {
		panic(err)
	}
	defer io.Stop()

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
