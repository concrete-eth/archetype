package rpc

import (
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/sim"
	"github.com/concrete-eth/archetype/testutils"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	chainId   = big.NewInt(1337)
	pcAddress = common.HexToAddress("0x1234")
)

func newTestSignerFn(t *testing.T) (common.Address, bind.SignerFn) {
	privateKeyHex := "b6caec81f24a057222a99f925671a845f5f27944e627e4097e5d7689b8981511"
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

	pc := precompile.NewCorePrecompile(specs, &testutils.TestCore{})
	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, pcAddress, pc)

	from, _ := newTestSignerFn(t)
	alloc := core.GenesisAlloc{from: {Balance: big.NewInt(1e18)}}

	return sim.NewSimulatedBackend(alloc, 1e8, registry)
}

func TestSendAction(t *testing.T) {
	var (
		specs  = testutils.NewTestArchSpecs(t)
		ethcli = newTestSimulatedBackend(t)
	)

	// Create a new action sender
	from, signerFn := newTestSignerFn(t)
	sender := NewActionSender(ethcli, specs.Actions, nil, pcAddress, from, 0, signerFn)

	// Send an action
	action := &testutils.ActionData_Add{}
	tx, err := sender.SendAction(action)
	if err != nil {
		t.Fatal(err)
	}

	// Commit the transaction
	ethcli.Commit()

	// Check the transaction receipt
	receipt, err := ethcli.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("expected status 1, got %d", receipt.Status)
	}
	if len(receipt.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(receipt.Logs))
	}

	// Check the log
	logAction, err := specs.Actions.LogToAction(*receipt.Logs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(logAction, action) {
		t.Fatalf("expected action, got %v", logAction)
	}
}

func waitForActionBatch(t *testing.T, actionBatchesChan <-chan archtypes.ActionBatch) archtypes.ActionBatch {
	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
		return archtypes.ActionBatch{}
	case actionBatchIn := <-actionBatchesChan:
		return actionBatchIn
	}
}

func TestSubscribeToActionBatches(t *testing.T) {
	var (
		specs  = testutils.NewTestArchSpecs(t)
		ethcli = newTestSimulatedBackend(t)
	)

	// Subscribe to action batches
	actionBatchesChan := make(chan archtypes.ActionBatch, 1)
	sub := SubscribeActionBatches(ethcli, specs.Actions, pcAddress, 0, actionBatchesChan)
	defer sub.Unsubscribe()

	// Commit and empty block
	ethcli.Commit()

	var batch archtypes.ActionBatch
	// Block 0
	if batch = waitForActionBatch(t, actionBatchesChan); batch.Len() != 0 {
		t.Fatalf("expected 0 actions, got %d", batch.Len())
	}
	// Block 1
	if batch = waitForActionBatch(t, actionBatchesChan); batch.Len() != 0 {
		t.Fatalf("expected 0 actions, got %d", batch.Len())
	}

	// Create a new action sender
	from, signerFn := newTestSignerFn(t)
	sender := NewActionSender(ethcli, specs.Actions, nil, pcAddress, from, 0, signerFn)

	// Send an action
	action := &testutils.ActionData_Add{}
	if _, err := sender.SendAction(action); err != nil {
		t.Fatal(err)
	}

	// Wait for the action batch
	ethcli.Commit()
	batch = waitForActionBatch(t, actionBatchesChan)

	// Check the action batch
	if batch.Len() != 1 {
		t.Fatalf("expected 1 action, got %d", batch.Len())
	}
	if !reflect.DeepEqual(batch.Actions[0], action) {
		t.Fatalf("expected action, got %v", batch.Actions[0])
	}
}
