package simulated

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var TickDepositorKeyHex = "6ad661dea8f20dff26c1529ec80eb479e16c39468dcacfe72d2a819ec59a4f25"

type TickingSimulatedBackend struct {
	*SimulatedBackend
	stopChan   chan struct{}
	tickOpts   *bind.TransactOpts
	tickTarget common.Address
	lock       sync.Mutex
}

// NewTickingSimulatedBackend creates a new simulated blockchain, pre-funded with the given accounts and with the given gas limit.
// It will automatically send a transaction to the tickTarget address every block once it is started.
func NewTickingSimulatedBackend(alloc types.GenesisAlloc, gasLimit uint64, concreteRegistry concrete.PrecompileRegistry) *TickingSimulatedBackend {
	tickPrivateKey, err := crypto.HexToECDSA(TickDepositorKeyHex)
	if err != nil {
		panic(err)
	}
	tickDepositionAddress := crypto.PubkeyToAddress(tickPrivateKey.PublicKey)
	alloc[tickDepositionAddress] = types.Account{Balance: math.MaxBig256}

	sim := NewSimulatedBackend(alloc, gasLimit, concreteRegistry)
	chainID := sim.Blockchain().Config().ChainID
	opts, err := bind.NewKeyedTransactorWithChainID(tickPrivateKey, chainID)
	if err != nil {
		panic(err)
	}
	opts.GasLimit = gasLimit / 2

	return &TickingSimulatedBackend{
		SimulatedBackend: sim,
		stopChan:         make(chan struct{}),
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
	tsb.insertTickTxIfNecessary()
	return tsb.SimulatedBackend.SendTransaction(ctx, tx)
}

func (tsb *TickingSimulatedBackend) Commit() {
	tsb.insertTickTxIfNecessary()
	tsb.SimulatedBackend.Commit()
}

func (tsb *TickingSimulatedBackend) insertTickTxIfNecessary() {
	tsb.lock.Lock()
	defer tsb.lock.Unlock()
	if tsb.tickTarget != (common.Address{}) && tsb.pendingBlock.Transactions().Len() == 0 {
		tickTx := tsb.newTickTransaction()
		if err := tsb.SimulatedBackend.SendTransaction(context.Background(), tickTx); err != nil {
			panic(err)
		}
	}
}

func (tsb *TickingSimulatedBackend) newTickTransaction() *types.Transaction {
	nonce, err := tsb.PendingNonceAt(context.Background(), tsb.tickOpts.From)
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
