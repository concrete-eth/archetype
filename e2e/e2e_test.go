package e2e

import (
	"math/big"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/precompile"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/concrete-eth/archetype/simulated"
	"github.com/concrete-eth/archetype/testutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	testChainId       = big.NewInt(1337)
	testPcAddress     = common.HexToAddress("0x1234")
	testPrivateKeyHex = "b6caec81f24a057222a99f925671a845f5f27944e627e4097e5d7689b8981511"
)

func newTestTxOpts(t *testing.T) *bind.TransactOpts {
	privateKey, err := crypto.HexToECDSA(testPrivateKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, testChainId)
	opts.Nonce = common.Big0
	if err != nil {
		t.Fatal(err)
	}
	return opts
}

func newTestSimulatedBackend(t *testing.T) *simulated.SimulatedBackend {
	schemas := testutils.NewTestArchSchemas(t)

	pc := precompile.NewCorePrecompile(schemas, func() arch.Core { return &testutils.Core{} })
	registry := concrete.NewRegistry()
	registry.AddPrecompile(0, testPcAddress, pc)

	from := newTestTxOpts(t).From
	alloc := types.GenesisAlloc{from: {Balance: big.NewInt(1e18)}}

	return simulated.NewSimulatedBackend(alloc, 1e8, registry)
}

func TestE2E(t *testing.T) {
	var (
		blockTime                  = 10 * time.Millisecond
		startingBlockNumber uint64 = 0
		kv                         = kvstore.NewMemoryKeyValueStore()
		core                       = &testutils.Core{}
	)
	var (
		ethcli  = newTestSimulatedBackend(t)
		auth    = newTestTxOpts(t)
		schemas = testutils.NewTestArchSchemas(t)
		io      = rpc.NewIO(ethcli, blockTime, schemas, auth, testPcAddress, testPcAddress, startingBlockNumber, 0)
		client  = io.NewClient(kv, core)
	)

	txUpdateChan := make(chan *rpc.ActionTxUpdate, 1)
	io.SetTxUpdateHook(func(txUpdate *rpc.ActionTxUpdate) {
		txUpdateChan <- txUpdate
	})

	ethcli.Commit()

	// Sync
	if err := client.SyncUntil(2); err != nil {
		t.Fatal(err)
	}

	// Send an action
	actionIn := &testutils.ActionData_Add{Summand: 1}
	if err := client.SendAction(actionIn); err != nil {
		t.Fatal(err)
	}
	// Wait for the transaction to be sent
	if u := <-txUpdateChan; u.Status != rpc.ActionTxStatus_Unsent {
		t.Fatalf("expected tx status to be pending, got %v", u.Status)
	}
	if u := <-txUpdateChan; u.Status != rpc.ActionTxStatus_Pending {
		t.Fatalf("expected tx status to be success, got %v", u.Status)
	}

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
	tableGetter := rpc.NewTableReader(ethcli, schemas.Tables, testPcAddress)
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
