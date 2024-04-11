package client

import (
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/kvstore"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/log"
)

type testCore struct {
	archtypes.BaseCore
}

type testAction struct{}

var _ archtypes.Core = (*testCore)(nil)

func (c *testCore) setVal(val uint64) {
	key := common.Hash{}
	bn := big.NewInt(int64(val))
	raw := bn.Bytes()
	c.KV().Set(key, common.BytesToHash(raw))
}

func (c *testCore) getVal() uint64 {
	key := common.Hash{}
	raw := c.KV().Get(key).Bytes()
	bn := new(big.Int).SetBytes(raw)
	return bn.Uint64()
}

func (c *testCore) double() {
	v := c.getVal()
	c.setVal(v * 2)
}

func (c *testCore) inc() {
	v := c.getVal()
	c.setVal(v + 1)
}

func (c *testCore) TicksPerBlock() uint {
	return 2
}

func (c *testCore) TestAction(action *testAction) error {
	c.inc()
	return nil
}

func (c *testCore) Tick() {
	c.double()
}

func newTestClient(t *testing.T) (*Client, lib.KeyValueStore, chan archtypes.ActionBatch, chan []archtypes.Action) {
	actionSpecs, err := archtypes.NewActionSpecsFromRaw(
		`[{"inputs":[],"name":"testAction","outputs":[],"stateMutability":"nonpayable","type":"function"}]`,
		`{"testAction":{"schema":{}}}`,
		map[string]reflect.Type{"TestAction": reflect.TypeOf(&testAction{})},
	)
	if err != nil {
		t.Fatal(err)
	}
	var (
		specs                    = archtypes.ArchSpecs{Actions: actionSpecs}
		core                     = &testCore{}
		kv                       = kvstore.NewMemoryKeyValueStore()
		actionBatchInChan        = make(chan archtypes.ActionBatch)
		actionOutChan            = make(chan []archtypes.Action)
		blockTime                = 10 * time.Millisecond
		blockNumber       uint64 = 0
	)
	client, err := New(specs, core, kv, actionBatchInChan, actionOutChan, blockTime, blockNumber)
	if err != nil {
		t.Fatal(err)
	}
	return client, kv, actionBatchInChan, actionOutChan
}

func TestSimulate(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	client.Simulate(func(_core archtypes.Core) {
		core := _core.(*testCore)
		core.inc()
		if core.getVal() != 1 {
			t.Fatal("unexpected value")
		}
	})
	if client.core.(*testCore).getVal() != 0 {
		t.Fatal("unexpected value")
	}
}

func TestSendActions(t *testing.T) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlError, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	client, _, _, actionOutChan := newTestClient(t)
	actionsIn := []archtypes.Action{&archtypes.CanonicalTickAction{}, &testAction{}}
	go client.SendActions(actionsIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionOutChan:
		if !reflect.DeepEqual(actionsIn, actionsOut) {
			t.Fatal("unexpected actions")
		}
	}
}

func TestSendAction(t *testing.T) {
	client, _, _, actionOutChan := newTestClient(t)
	actionIn := &testAction{}
	go client.SendAction(actionIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionOutChan:
		if len(actionsOut) != 1 {
			t.Fatal("unexpected actions")
		}
		if !reflect.DeepEqual(actionIn, actionsOut[0]) {
			t.Fatal("unexpected action")
		}
	}
}

var testData = []struct {
	batch                     archtypes.ActionBatch
	expectedTickActionInBatch bool
	expectedCoreVal           uint64
}{
	{
		archtypes.ActionBatch{
			BlockNumber: 0,
			Actions: []archtypes.Action{
				&archtypes.CanonicalTickAction{},
				&testAction{},
			},
		},
		true,
		1,
	},
	{
		archtypes.ActionBatch{
			BlockNumber: 1,
			Actions: []archtypes.Action{
				&archtypes.CanonicalTickAction{},
			},
		},
		true,
		4,
	},
	{
		archtypes.ActionBatch{
			BlockNumber: 2,
			Actions: []archtypes.Action{
				&testAction{},
			},
		},
		false,
		5,
	},
	{
		archtypes.ActionBatch{
			BlockNumber: 3,
			Actions:     []archtypes.Action{},
		},
		false,
		5,
	},
}

func init() {
	for i, data := range testData {
		if data.batch.BlockNumber != uint64(i) {
			panic("unexpected block number")
		}
	}
}

func TestSync(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	actionBatchInChan := make(chan archtypes.ActionBatch, 1)
	client.actionBatchInChan = actionBatchInChan

	didReceiveNewBatch, didTick, err := client.Sync()
	if err != nil {
		t.Fatal(err)
	}
	if didReceiveNewBatch {
		t.Fatal("unexpected new batch")
	}
	if didTick {
		t.Fatal("expected no tick action")
	}
	if client.core.BlockNumber() != 0 {
		t.Fatal("unexpected block number")
	}
	if client.core.(*testCore).getVal() != 0 {
		t.Fatal("unexpected value")
	}

	for _, actionBatch := range testData {
		actionBatchInChan <- actionBatch.batch
		_, didTick, err = client.Sync()
		if err != nil {
			t.Fatal(err)
		}
		if didTick != actionBatch.expectedTickActionInBatch {
			t.Fatal("unexpected tick action")
		}
		if client.core.BlockNumber() != actionBatch.batch.BlockNumber+1 {
			t.Fatal("unexpected block number")
		}
		if client.core.(*testCore).getVal() != actionBatch.expectedCoreVal {
			t.Fatal("unexpected value")
		}
	}
}

func TestSyncUntil(t *testing.T) {
	client, _, actionBatchInChan, _ := newTestClient(t)

	err := client.SyncUntil(0)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _, actionBatch := range testData {
			actionBatchInChan <- actionBatch.batch
		}
	}()

	for _, blockToSyncTo := range []uint64{2, 4} {
		err := client.SyncUntil(blockToSyncTo)
		if err != nil {
			t.Fatal(err)
		}
		if client.core.BlockNumber() != blockToSyncTo {
			t.Fatal("unexpected block number")
		}
		if client.core.(*testCore).getVal() != testData[blockToSyncTo-1].expectedCoreVal {
			t.Fatal("unexpected value")
		}
	}
}

func TestInterpolatedSync(t *testing.T) {
	client, _, actionBatchInChan, _ := newTestClient(t)

	didReceiveNewBatch, didTick, err := client.InterpolatedSync()
	if err != nil {
		t.Fatal(err)
	}
	if didReceiveNewBatch {
		t.Fatal("unexpected new batch")
	}
	if !didTick {
		t.Fatal("expected tick")
	}

	go func() {
		ticker := time.NewTicker(client.blockTime)
		defer ticker.Stop()
		for _, data := range testData {
			<-ticker.C
			actionBatchInChan <- data.batch
		}
	}()

	var (
		ticksPerBlock = client.core.TicksPerBlock()
		tickPeriod    = client.blockTime / time.Duration(ticksPerBlock)
	)

	ticker := time.NewTicker(client.blockTime / 4)

	for range ticker.C {
		didReceiveNewBatch, _, err := client.InterpolatedSync()
		if err != nil {
			t.Fatal(err)
		}
		var expectedCoreVal uint64
		if client.core.BlockNumber() == 0 {
			expectedCoreVal = 0
		} else {
			expectedCoreVal = testData[client.core.BlockNumber()-1].expectedCoreVal
		}
		if !didReceiveNewBatch {
			// Adjust expectedCoreVal for interpolated ticks
			targetTicks := uint(time.Since(client.lastNewBatchTime)/tickPeriod) + 1
			targetTicks = utils.Min(targetTicks, ticksPerBlock)
			expectedCoreVal *= uint64(2 * targetTicks)
		}
		if client.core.(*testCore).getVal() != expectedCoreVal {
			t.Fatal("unexpected value")
		}
		if int(client.core.BlockNumber()) >= len(testData) {
			break
		}
	}

	if client.core.(*testCore).getVal() != testData[len(testData)-1].expectedCoreVal {
		t.Fatal("unexpected value")
	}
}
