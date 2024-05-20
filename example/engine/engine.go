package engine

import (
	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/physics"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/snapshot"
	"github.com/ethereum/go-ethereum/cmd/geth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	concrete_rpc "github.com/ethereum/go-ethereum/concrete/rpc"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

const (
	SnapshotNamespace = "arch"
)

func NewRegistry() concrete.PrecompileRegistry {
	schemas := arch.ArchSchemas{Actions: archmod.ActionSchemas, Tables: archmod.TableSchemas}
	pc := precompile.NewCorePrecompile(schemas, func() arch.Core { return &physics.Core{} })
	address := common.HexToAddress("0x80")
	startingBlock := uint64(0)
	registry := concrete.NewRegistry()
	registry.AddPrecompile(startingBlock, address, pc)
	return registry
}

func SnapshotWriterConstructor(ethereum *eth.Ethereum) rpc.API {
	return rpc.API{
		Namespace:     SnapshotNamespace,
		Authenticated: true,
		Service:       snapshot.Root.NewWriter(ethereum),
	}
}

var _ concrete_rpc.APIConstructor = SnapshotWriterConstructor

func SnapshotReaderConstructor(ethereum *eth.Ethereum) rpc.API {
	return rpc.API{
		Namespace:     SnapshotNamespace,
		Authenticated: false,
		Service:       snapshot.Root.NewReader(ethereum),
	}
}

var _ concrete_rpc.APIConstructor = SnapshotReaderConstructor

func NewGeth() *cli.App {
	registry := NewRegistry()
	apis := []concrete_rpc.APIConstructor{
		SnapshotWriterConstructor,
		SnapshotReaderConstructor,
	}
	go snapshot.Root.RunSnapshotWorker()
	if snapshot.Root.IsSchedulerEnabled() {
		go snapshot.Root.RunScheduler()
	}
	return geth.NewConcreteGethApp(registry, apis)
}
