package client

import (
	"errors"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/kvstore"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type testCore struct {
	kv          lib.KeyValueStore
	blockNumber BlockNumber
	tickIndex   uint
}

type testAction struct{}

var _ Core = (*testCore)(nil)

func (c *testCore) setVal(val uint64) {
	key := common.Hash{}
	bn := big.NewInt(int64(val))
	raw := bn.Bytes()
	c.kv.Set(key, common.BytesToHash(raw))
}

func (c *testCore) getVal() uint64 {
	key := common.Hash{}
	raw := c.kv.Get(key).Bytes()
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

func (c *testCore) SetKV(kv lib.KeyValueStore) error {
	c.kv = kv
	return nil
}

func (c *testCore) ExecuteAction(action archtypes.Action) error {
	switch action.(type) {
	case *archtypes.CanonicalTickAction:
		c.RunBlockTicks()
	case *testAction:
		c.inc()
	default:
		return errors.New("unexpected action")
	}
	return nil
}

func (c *testCore) SetBlockNumber(blockNumber BlockNumber) {
	c.blockNumber = blockNumber
}

func (c *testCore) BlockNumber() BlockNumber {
	return c.blockNumber
}

func (c *testCore) RunSingleTick() {
	c.double()
}

func (c *testCore) RunBlockTicks() {
	for i := uint(0); i < c.TicksPerBlock(); i++ {
		c.RunSingleTick()
	}
}

func (c *testCore) TicksPerBlock() uint {
	return 2
}

func (c *testCore) ExpectTick() bool {
	return true
}

func (c *testCore) SetInBlockTickIndex(index uint) {
	c.tickIndex = index
}

func (c *testCore) InBlockTickIndex() uint {
	return c.tickIndex
}

func newTestClient(t *testing.T) (*Client, lib.KeyValueStore, chan archtypes.ActionBatch, chan []archtypes.Action) {
	var (
		core              = &testCore{}
		kv                = kvstore.NewMemoryKeyValueStore()
		actionBatchInChan = make(chan archtypes.ActionBatch)
		actionOutChan     = make(chan []archtypes.Action)
		blockTime         = 10 * time.Millisecond
		blockNumber       = archtypes.NewBlockNumber(0)
	)
	client, err := New(core, kv, actionBatchInChan, actionOutChan, blockTime, blockNumber)
	if err != nil {
		t.Fatal(err)
	}
	return client, kv, actionBatchInChan, actionOutChan
}

func TestSimulate(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	client.Simulate(func(_core Core) {
		core := _core.(*testCore)
		core.inc()
		if core.getVal() != 1 {
			t.Fatal("unexpected value")
		}
	})
	if client.Core.(*testCore).getVal() != 0 {
		t.Fatal("unexpected value")
	}
}

func TestSendActions(t *testing.T) {
	client, _, _, actionOutChan := newTestClient(t)
	actionsIn := []archtypes.Action{&testAction{}, &testAction{}}
	go client.SendActions(actionsIn)
	select {
	case <-time.After(20 * time.Millisecond):
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
			BlockNumber: archtypes.NewBlockNumber(0),
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
			BlockNumber: archtypes.NewBlockNumber(1),
			Actions: []archtypes.Action{
				&archtypes.CanonicalTickAction{},
			},
		},
		true,
		4,
	},
	{
		archtypes.ActionBatch{
			BlockNumber: archtypes.NewBlockNumber(2),
			Actions: []archtypes.Action{
				&testAction{},
			},
		},
		false,
		5,
	},
	{
		archtypes.ActionBatch{
			BlockNumber: archtypes.NewBlockNumber(3),
			Actions:     []archtypes.Action{},
		},
		false,
		5,
	},
}

// TODO: use uint64 for block number

func init() {
	for i, data := range testData {
		if data.batch.BlockNumber != archtypes.NewBlockNumber(uint64(i)) {
			panic("unexpected block number")
		}
	}
}

func TestSync(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	actionBatchInChan := make(chan archtypes.ActionBatch, 1)
	client.actionBatchInChan = actionBatchInChan

	_, tickActionInBatch, err := client.Sync() // TODO
	if err != nil {
		t.Fatal(err)
	}
	if tickActionInBatch {
		t.Fatal("expected no tick action")
	}
	if client.Core.BlockNumber().Uint64() != 0 {
		t.Fatal("unexpected block number")
	}
	if client.Core.(*testCore).getVal() != 0 {
		t.Fatal("unexpected value")
	}

	for _, actionBatch := range testData {
		actionBatchInChan <- actionBatch.batch
		_, tickActionInBatch, err = client.Sync()
		if err != nil {
			t.Fatal(err)
		}
		if tickActionInBatch != actionBatch.expectedTickActionInBatch {
			t.Fatal("unexpected tick action")
		}
		if client.Core.BlockNumber().Uint64() != actionBatch.batch.BlockNumber.AddUint64(1).Uint64() {
			t.Fatal("unexpected block number")
		}
		if client.Core.(*testCore).getVal() != actionBatch.expectedCoreVal {
			t.Fatal("unexpected value")
		}
	}
}

func TestSyncUntil(t *testing.T) {
	client, _, actionBatchInChan, _ := newTestClient(t)

	err := client.SyncUntil(archtypes.NewBlockNumber(0))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _, actionBatch := range testData {
			actionBatchInChan <- actionBatch.batch
		}
	}()

	for _, blockToSyncTo := range []uint64{2, 4} {
		err := client.SyncUntil(archtypes.NewBlockNumber(blockToSyncTo))
		if err != nil {
			t.Fatal(err)
		}
		if client.Core.BlockNumber().Uint64() != blockToSyncTo {
			t.Fatal("unexpected block number")
		}
		if client.Core.(*testCore).getVal() != testData[blockToSyncTo-1].expectedCoreVal {
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
		ticksPerBlock = client.Core.TicksPerBlock()
		tickPeriod    = client.blockTime / time.Duration(ticksPerBlock)
	)

	ticker := time.NewTicker(client.blockTime / 4)

	for range ticker.C {
		didReceiveNewBatch, _, err := client.InterpolatedSync()
		if err != nil {
			t.Fatal(err)
		}
		var expectedCoreVal uint64
		if client.Core.BlockNumber().Uint64() == 0 {
			expectedCoreVal = 0
		} else {
			expectedCoreVal = testData[client.Core.BlockNumber().Uint64()-1].expectedCoreVal
		}
		if !didReceiveNewBatch {
			// Adjust expectedCoreVal for interpolated ticks
			targetTicks := uint(time.Since(client.lastNewBatchTime)/tickPeriod) + 1
			targetTicks = utils.Min(targetTicks, ticksPerBlock)
			expectedCoreVal *= uint64(2 * targetTicks)
		}
		if client.Core.(*testCore).getVal() != expectedCoreVal {
			t.Fatal("unexpected value")
		}
		if int(client.Core.BlockNumber().Uint64()) >= len(testData) {
			break
		}
	}

	if client.Core.(*testCore).getVal() != testData[len(testData)-1].expectedCoreVal {
		t.Fatal("unexpected value")
	}
}
