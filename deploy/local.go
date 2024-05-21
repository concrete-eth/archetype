package deploy

import (
	"math/big"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/concrete-eth/archetype/simulated"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var (
	localChainId       = big.NewInt(1337)
	localPrivateKeyHex = "504d29ac79864050983ca646570b0bbe158fa5878c1bda7f1fdb0a48bd8b37b6"
)

func NewSimulatedBackend(registry concrete.PrecompileRegistry, gasLimit uint64, devAddresses ...common.Address) *simulated.TickingSimulatedBackend {
	alloc := types.GenesisAlloc{}
	for _, addr := range devAddresses {
		alloc[addr] = types.Account{Balance: math.MaxBig256}
	}
	return simulated.NewTickingSimulatedBackend(alloc, gasLimit, registry)
}

func NewLocalIO(registry concrete.PrecompileRegistry, schemas arch.ArchSchemas, deployer GameContractDeployer, logic common.Address, blockTime time.Duration) (*rpc.IO, error) {
	// Load tx opts
	privateKey, err := crypto.HexToECDSA(localPrivateKeyHex)
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, localChainId)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(0)

	// Create simulated backend with precompile
	ethcli := NewSimulatedBackend(registry, 100_000_000, auth.From)
	// Deploy game
	gameAddr, coreAddr, err := DeployGame(auth, ethcli, deployer, logic, true)
	if err != nil {
		return nil, err
	}
	ethcli.Commit()

	// Create chain IO
	io := rpc.NewIO(ethcli, blockTime, schemas, auth, gameAddr, coreAddr, 0, 0)
	io.SetTxUpdateHook(func(txUpdate *rpc.ActionTxUpdate) {
		log.Info("Transaction "+txUpdate.Status.String(), "nonce", txUpdate.Nonce, "txHash", txUpdate.TxHash.Hex())
	})

	// Start simulated ticking
	ethcli.Start(blockTime, gameAddr)

	io.RegisterCancelFn(func() {
		// Stop ticking when the IO is closed
		defer ethcli.Stop()
	})

	return io, nil
}
