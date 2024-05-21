package client

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/testutils"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

func newTestClient(t *testing.T) (*Client, lib.KeyValueStore, chan arch.ActionBatch, chan []arch.Action) {
	var (
		schemas                = testutils.NewTestArchSchemas(t)
		core                   = &testutils.Core{}
		kv                     = kvstore.NewMemoryKeyValueStore()
		actionBatchChan        = make(chan arch.ActionBatch)
		actionChan             = make(chan []arch.Action)
		blockTime              = 1 * time.Second
		blockNumber     uint64 = 0
	)
	client := New(schemas, core, kv, actionBatchChan, actionChan, blockTime, blockNumber)
	return client, kv, actionBatchChan, actionChan
}

func TestSimulate(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	client.Simulate(func(_core arch.Core) {
		core := _core.(*testutils.Core)
		core.SetCounter(1)
		if c := core.GetCounter(); c != 1 {
			t.Errorf("expected %v, got %v", 1, c)
		}
	})
	if c := client.Core().(*testutils.Core).GetCounter(); c != 0 {
		t.Errorf("expected %v, got %v", 0, c)
	}
}

func TestSendActions(t *testing.T) {
	client, _, _, actionChan := newTestClient(t)
	actionsIn := []arch.Action{&arch.CanonicalTickAction{}, &testutils.ActionData_Add{}}
	go client.SendActions(actionsIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionChan:
		if !reflect.DeepEqual(actionsIn, actionsOut) {
			t.Fatal("unexpected actions")
		}
	}
}

func TestSendAction(t *testing.T) {
	client, _, _, actionChan := newTestClient(t)
	actionIn := &testutils.ActionData_Add{}
	go client.SendAction(actionIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionChan:
		if len(actionsOut) != 1 {
			t.Fatal("unexpected actions")
		}
		if !reflect.DeepEqual(actionIn, actionsOut[0]) {
			t.Fatal("unexpected action")
		}
	}
}

var testData = []struct {
	batch                arch.ActionBatch
	expTickActionInBatch bool
	expCounterValue      int16
}{
	{
		arch.ActionBatch{
			BlockNumber: 0,
			Actions: []arch.Action{
				&arch.CanonicalTickAction{},
				&testutils.ActionData_Add{Summand: 1},
			},
		},
		true,
		1,
	},
	{
		arch.ActionBatch{
			BlockNumber: 1,
			Actions: []arch.Action{
				&arch.CanonicalTickAction{},
			},
		},
		true,
		4,
	},
	{
		arch.ActionBatch{
			BlockNumber: 2,
			Actions: []arch.Action{
				&testutils.ActionData_Add{Summand: 1},
			},
		},
		false,
		5,
	},
	{
		arch.ActionBatch{
			BlockNumber: 3,
			Actions:     []arch.Action{},
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
	actionBatchChan := make(chan arch.ActionBatch, 1)
	client.actionBatchInChan = actionBatchChan

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
	if client.Core().BlockNumber() != 0 {
		t.Fatal("unexpected block number")
	}
	if client.Core().(*testutils.Core).GetCounter() != 0 {
		t.Fatal("unexpected value")
	}

	for _, actionBatch := range testData {
		actionBatchChan <- actionBatch.batch
		_, didTick, err = client.Sync()
		if err != nil {
			t.Fatal(err)
		}
		if didTick != actionBatch.expTickActionInBatch {
			t.Errorf("expected %v, got %v", actionBatch.expTickActionInBatch, didTick)
		}
		if client.Core().BlockNumber() != actionBatch.batch.BlockNumber+1 {
			t.Errorf("expected %v, got %v", actionBatch.batch.BlockNumber+1, client.Core().BlockNumber())
		}
		if c := client.Core().(*testutils.Core).GetCounter(); c != actionBatch.expCounterValue {
			t.Errorf("expected %v, got %v", actionBatch.expCounterValue, c)
		}
	}
}

func TestSyncUntil(t *testing.T) {
	client, _, actionBatchChan, _ := newTestClient(t)

	err := client.SyncUntil(0)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _, actionBatch := range testData {
			actionBatchChan <- actionBatch.batch
		}
	}()

	for _, blockToSyncTo := range []uint64{2, 4} {
		err := client.SyncUntil(blockToSyncTo)
		if err != nil {
			t.Fatal(err)
		}
		if client.Core().BlockNumber() != blockToSyncTo {
			t.Fatal("unexpected block number")
		}
		if c := client.Core().(*testutils.Core).GetCounter(); c != testData[blockToSyncTo-1].expCounterValue {
			t.Errorf("expected %v, got %v", testData[blockToSyncTo-1].expCounterValue, c)
		}
	}
}

type testClock struct {
	t time.Time
}

func (c *testClock) Now() time.Time {
	return c.t
}

func (c *testClock) Advance(d time.Duration) {
	c.t = c.t.Add(d)
}

func TestInterpolatedSync(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	actionBatchChan := make(chan arch.ActionBatch, 1)
	client.actionBatchInChan = actionBatchChan
	clock := &testClock{}
	client.now = clock.Now
	startTime := clock.Now()

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

	var (
		ticksPerBlock = client.Core().TicksPerBlock()
		tickPeriod    = client.blockTime / time.Duration(ticksPerBlock)
	)

	idx := 0
	for {
		var expectReceiveNewBatch bool
		if clock.Now().Sub(startTime) >= time.Duration(idx)*client.blockTime {
			if idx >= len(testData) {
				break
			}
			actionBatchChan <- testData[idx].batch
			expectReceiveNewBatch = true
			idx++
		}

		didReceiveNewBatch, _, err := client.InterpolatedSync()
		if err != nil {
			t.Fatal(err)
		}
		if didReceiveNewBatch != expectReceiveNewBatch {
			t.Errorf("expected %v, got %v", expectReceiveNewBatch, didReceiveNewBatch)
		}

		var expectedCoreVal int16
		if client.Core().BlockNumber() == 0 {
			expectedCoreVal = 0
		} else {
			expectedCoreVal = testData[client.Core().BlockNumber()-1].expCounterValue
		}

		// Adjust expectedCoreVal for interpolated ticks
		targetTicks := uint64(clock.Now().Sub(client.lastNewBatchTime)/tickPeriod) + 1
		targetTicks = utils.Min(targetTicks, ticksPerBlock)
		expectedCoreVal *= int16(math.Pow(2, float64(targetTicks)))

		if client.core.InBlockTickIndex() != targetTicks-1 {
			t.Errorf("expected %v, got %v", targetTicks, client.core.InBlockTickIndex())
		}
		if c := client.Core().(*testutils.Core).GetCounter(); c != expectedCoreVal {
			t.Errorf("expected %v, got %v", expectedCoreVal, c)
		}

		clock.Advance(client.blockTime / time.Duration(ticksPerBlock*2))
	}
}
