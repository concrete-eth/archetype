package simulated

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

type SimulatedBackend struct {
	*backends.SimulatedBackend
	db ethdb.Database
}

// NewSimulatedBackend creates a new simulated blockchain, pre-funded with the given accounts and with the given gas limit.
func NewSimulatedBackend(alloc core.GenesisAlloc, gasLimit uint64, concreteRegistry concrete.PrecompileRegistry) *SimulatedBackend {
	return NewSimulatedBackendWithDatabase(rawdb.NewMemoryDatabase(), alloc, gasLimit, concreteRegistry)
}

// NewSimulatedBackendWithDatabase creates a new simulated blockchain, pre-funded with the given accounts and with the given gas limit.
func NewSimulatedBackendWithDatabase(db ethdb.Database, alloc core.GenesisAlloc, gasLimit uint64, concreteRegistry concrete.PrecompileRegistry) *SimulatedBackend {
	sim := &SimulatedBackend{
		SimulatedBackend: backends.NewSimulatedBackendWithDatabase(db, alloc, gasLimit),
		db:               db,
	}
	sim.Blockchain().SetConcrete(concreteRegistry)
	return sim
}

func (sb *SimulatedBackend) BlockChain() *core.BlockChain {
	return sb.Blockchain()
}

func (sb *SimulatedBackend) BlockNumber(_ context.Context) (uint64, error) {
	return sb.Blockchain().CurrentBlock().Number.Uint64(), nil
}

func (sb *SimulatedBackend) ChainDb() ethdb.Database {
	return sb.db
}

var TickDepositorKeyHex = "6ad661dea8f20dff26c1529ec80eb479e16c39468dcacfe72d2a819ec59a4f25"

type TickingSimulatedBackend struct {
	*SimulatedBackend
	stopChan   chan struct{}
	txQueue    []*types.Transaction
	tickOpts   *bind.TransactOpts
	tickTarget common.Address
}

// NewTickingSimulatedBackend creates a new simulated blockchain, pre-funded with the given accounts and with the given gas limit.
// It will automatically send a transaction to the tickTarget address every block once it is started.
func NewTickingSimulatedBackend(alloc core.GenesisAlloc, gasLimit uint64, concreteRegistry concrete.PrecompileRegistry) *TickingSimulatedBackend {
	tickPrivateKey, err := crypto.HexToECDSA(TickDepositorKeyHex)
	if err != nil {
		panic(err)
	}
	tickDepositionAddress := crypto.PubkeyToAddress(tickPrivateKey.PublicKey)
	alloc[tickDepositionAddress] = core.GenesisAccount{Balance: math.MaxBig256}

	sim := NewSimulatedBackend(alloc, gasLimit, concreteRegistry)
	chainID := sim.BlockChain().Config().ChainID
	opts, err := bind.NewKeyedTransactorWithChainID(tickPrivateKey, chainID)
	if err != nil {
		panic(err)
	}
	opts.GasLimit = gasLimit / 2

	return &TickingSimulatedBackend{
		SimulatedBackend: sim,
		stopChan:         make(chan struct{}),
		txQueue:          make([]*types.Transaction, 0),
		tickOpts:         opts,
	}
}

func (tsb *TickingSimulatedBackend) Start(blockTime time.Duration, tickTarget common.Address) {
	tsb.tickTarget = tickTarget
	go func() {
		ticker := time.NewTicker(blockTime)
		for {
			select {
			case <-tsb.stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				tsb.Commit()
			}
		}
	}()
}

func (tsb *TickingSimulatedBackend) Stop() {
	tsb.stopChan <- struct{}{}
}

func (tsb *TickingSimulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	tsb.txQueue = append(tsb.txQueue, tx)
	return nil
}

func (tsb *TickingSimulatedBackend) Commit() {
	if tsb.tickTarget != (common.Address{}) {
		tickTx := tsb.newTickTransaction()
		err := tsb.SimulatedBackend.SendTransaction(context.Background(), tickTx)
		if err != nil {
			panic(err)
		}
	}
	for _, tx := range tsb.txQueue {
		err := tsb.SimulatedBackend.SendTransaction(context.Background(), tx)
		if err != nil {
			panic(err)
		}
	}
	tsb.txQueue = make([]*types.Transaction, 0)
	tsb.SimulatedBackend.Commit()
}

func (tsb *TickingSimulatedBackend) newTickTransaction() *types.Transaction {
	nonce, err := tsb.NonceAt(context.Background(), tsb.tickOpts.From, nil)
	if err != nil {
		panic(err)
	}
	gasPrice, err := tsb.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	data := crypto.Keccak256([]byte("tick()"))[:4]
	tx := types.NewTransaction(nonce, tsb.tickTarget, common.Big0, tsb.tickOpts.GasLimit, gasPrice, data)
	tx, err = tsb.tickOpts.Signer(tsb.tickOpts.From, tx)
	if err != nil {
		panic(err)
	}
	return tx
}
