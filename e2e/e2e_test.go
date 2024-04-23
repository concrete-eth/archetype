package e2e

import (
	"math/big"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/client"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/concrete-eth/archetype/sim"
	"github.com/concrete-eth/archetype/testutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	chainId       = big.NewInt(1337)
	pcAddress     = common.HexToAddress("0x1234")
	privateKeyHex = "b6caec81f24a057222a99f925671a845f5f27944e627e4097e5d7689b8981511"
)

func newTestClient(t *testing.T) (*client.Client, lib.KeyValueStore, chan arch.ActionBatch, chan []arch.Action) {
	var (
		specs                    = testutils.NewTestArchSpecs(t)
		core                     = &testutils.Core{}
		kv                       = kvstore.NewMemoryKeyValueStore()
		actionBatchInChan        = make(chan arch.ActionBatch)
		actionOutChan            = make(chan []arch.Action)
		blockTime                = 10 * time.Millisecond
		blockNumber       uint64 = 0
	)
	client := client.New(specs, core, kv, actionBatchInChan, actionOutChan, blockTime, blockNumber)
	return client, kv, actionBatchInChan, actionOutChan
}

func newTestSignerFn(t *testing.T) (common.Address, bind.SignerFn) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		t.Fatal(err)
	}
	return opts.From, opts.Signer
}

func newTestSimulatedBackend(t *testing.T) *sim.SimulatedBackend {
	specs := testutils.NewTestArchSpecs(t)

	pc := precompile.NewCorePrecompile(specs, &testutils.Core{})
	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, pcAddress, pc)

	from, _ := newTestSignerFn(t)
	alloc := core.GenesisAlloc{from: {Balance: big.NewInt(1e18)}}

	return sim.NewSimulatedBackend(alloc, 1e8, registry)
}

func TestE2E(t *testing.T) {
	var (
		specs                                       = testutils.NewTestArchSpecs(t)
		client, _, actionBatchInChan, actionOutChan = newTestClient(t)
		ethcli                                      = newTestSimulatedBackend(t)
		txOutChan                                   = make(chan *rpc.ActionTxUpdate)
	)

	// Subscribe to action batches
	sub := rpc.SubscribeActionBatches(ethcli, specs.Actions, pcAddress, 0, actionBatchInChan)
	defer sub.Unsubscribe()

	// Create a new action sender
	from, signerFn := newTestSignerFn(t)
	sender := rpc.NewActionSender(ethcli, specs.Actions, nil, pcAddress, from, 0, signerFn)
	sender.StartSendingActions(actionOutChan, txOutChan)

	ethcli.Commit()

	if err := client.SyncUntil(2); err != nil {
		t.Fatal(err)
	}

	// Send an action
	actionIn := &testutils.ActionData_Add{Summand: 1}
	if err := client.SendAction(actionIn); err != nil {
		t.Fatal(err)
	}
	// Wait for the transaction to be sent
	<-txOutChan
	<-txOutChan

	ethcli.Commit()

	// Wait for the action batch
	timeout := time.After(10 * time.Millisecond)
	for {
		didReceiveNewBatch, didTick, err := client.Sync()
		if err != nil {
			t.Fatal(err)
		}
		if didTick {
			t.Fatal("expected no tick")
		}
		if didReceiveNewBatch {
			break
		}
		select {
		case <-timeout:
			t.Fatal("timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Read the counter
	localCounter := client.Core().(*testutils.Core).GetCounter()
	if localCounter != 1 {
		t.Errorf("expected local counter to be 1, got %d", localCounter)
	}

	// Read the counter from the chain
	tableGetter := rpc.NewTableReader(ethcli, specs.Tables, pcAddress)
	_remoteCounter, err := tableGetter.Read("Counter")
	if err != nil {
		t.Fatal(err)
	}
	remoteCounter, ok := _remoteCounter.(*testutils.RowData_Counter)
	if !ok {
		t.Fatalf("expected counter row, got %T", _remoteCounter)
	}
	if remoteCounter.GetValue() != 1 {
		t.Errorf("expected remote counter to be 1, got %d", remoteCounter)
	}
}
