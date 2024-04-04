package client

import (
	"errors"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/log"

	"github.com/concrete-eth/archetype/kvstore"
)

var (
	ErrBlockNumberMismatch = errors.New("block number mismatch")
	ErrTickNotFirst        = errors.New("tick not first action")
)

type BlockNumber uint64

func (b BlockNumber) Uint64() uint64 { return uint64(b) }

type Core interface {
	SetKV(kv lib.KeyValueStore) error // Set the key-value store
	ExecuteAction(Action) error       // Execute the given action
	SetBlockNumber(BlockNumber)       // Set the block number
	BlockNumber() BlockNumber         // Get the block number
	RunSingleTick()                   // Run a single tick
	RunBlockTicks()                   // Run all ticks in a block
	TicksPerBlock() uint              // Get the number of ticks per block
	ExpectTick() bool                 // Check if a tick is expected
	SetInBlockTickIndex(uint)         // Set the in-block tick index
	InBlockTickIndex() uint           // Get the in-block tick index
}

type Action interface{}

type CanonicalTickAction struct{}

// Holds all the actions included to a specific core in a specific block
type ActionBatch struct {
	BlockNumber BlockNumber
	Actions     []Action
}

type Client struct {
	Core Core
	kv   *kvstore.StagedKeyValueStore

	actionBatchInChan <-chan ActionBatch
	actionOutChan     chan<- []Action

	blockTime time.Duration

	lastNewBatchTime  time.Time
	ticksRunThisBlock uint

	lock sync.Mutex
}

func New(core Core, kv lib.KeyValueStore, actionBatchInChan <-chan ActionBatch, actionOutChan chan<- []Action, blockTime time.Duration, blockNumber BlockNumber) (*Client, error) {
	stagedKv := kvstore.NewStagedKeyValueStore(kv)
	if err := core.SetKV(stagedKv); err != nil {
		return nil, err
	}
	return &Client{
		Core:              core,
		kv:                stagedKv,
		actionBatchInChan: actionBatchInChan,
		actionOutChan:     actionOutChan,
		blockTime:         blockTime,
		ticksRunThisBlock: 0,
	}, nil
}

func (c *Client) debug(msg string, ctx ...interface{}) {
	log.Debug(msg, ctx...)
}

func (c *Client) warn(msg string, ctx ...interface{}) {
	log.Warn(msg, ctx...)
}

func (c *Client) error(msg string, ctx ...interface{}) {
	log.Error(msg, ctx...)
}

// Applies the given action batch to the core and returns whether a tick action was included in the batch
// If a tick action is included, it must be the first action in the batch
func (c *Client) applyBatch(batch ActionBatch) (bool, error) {
	if c.Core.BlockNumber() != batch.BlockNumber {
		return false, ErrBlockNumberMismatch
	}
	tickActionInBatch := false
	for ii, action := range batch.Actions {
		if err := c.Core.ExecuteAction(action); err != nil {
			c.error("failed to execute action", "err", err)
		}
		if _, ok := action.(*CanonicalTickAction); ok {
			if ii != 0 {
				return false, ErrTickNotFirst
			}
			tickActionInBatch = true
		}
	}
	return tickActionInBatch, nil
}

// Applies the given action batch to the core, commits the changes to the key-value store, and updates the core block number
func (c *Client) applyBatchAndCommit(batch ActionBatch) (bool, error) {
	tickActionInBatch, err := c.applyBatch(batch)
	if err != nil {
		return false, err
	}
	c.kv.Commit()
	c.lastNewBatchTime = time.Now()
	c.Core.SetBlockNumber(batch.BlockNumber + 1)
	return tickActionInBatch, nil
}

// Runs the given function and then reverts all the changes to the key-value store
func (c *Client) Simulate(f func(core Core)) {
	// Put another stage on top of the current key-value store that will never be committed
	// and will be discarded after the function is executed
	c.lock.Lock()
	defer c.lock.Unlock()
	simKv := kvstore.NewStagedKeyValueStore(c.kv)
	c.Core.SetKV(simKv)
	f(c.Core)
	c.Core.SetKV(c.kv)
}

func (c *Client) SendAction(action Action) {
	c.SendActions([]Action{action})
}

func (c *Client) SendActions(actions []Action) {
	actionsToSend := make([]Action, 0)
	c.Simulate(func(core Core) {
		for _, action := range actions {
			if err := core.ExecuteAction(action); err != nil {
				c.error("failed to execute action", "err", err)
				continue
			}
			actionsToSend = append(actionsToSend, action)
		}
	})

	c.actionOutChan <- actionsToSend
}

func (c *Client) Sync() (didReceiveNewBatch bool, didTick bool, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	select {
	case batch := <-c.actionBatchInChan:
		didTick, err := c.applyBatchAndCommit(batch)
		return true, didTick, err
	default:
		return false, false, nil
	}
}

func (c *Client) SyncUntil(blockNumber BlockNumber) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for {
		if c.Core.BlockNumber() >= blockNumber {
			break
		}
		batch, ok := <-c.actionBatchInChan
		if !ok {
			return errors.New("channel closed")
		}
		if _, err := c.applyBatchAndCommit(batch); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) InterpolatedSync() (didReceiveNewBatch bool, didTick bool, err error) {
	if !c.Core.ExpectTick() {
		return c.Sync()
	}
	select {
	case batch := <-c.actionBatchInChan:
		// Received a new batch of actions
		// Revert any tick anticipation and apply batch normally
		didReceiveNewBatch = true
		c.kv.Revert()
		didTick, err := c.applyBatchAndCommit(batch)
		c.ticksRunThisBlock = 0
		return didReceiveNewBatch, didTick, err
	default:
		// No new batch of actions received
		// Anticipate ticks corresponding to the expected tick action in the next block
		didReceiveNewBatch = false

		var (
			ticksPerBlock = c.Core.TicksPerBlock()
			tickPeriod    = c.blockTime / time.Duration(ticksPerBlock)
		)

		targetTicks := uint(time.Since(c.lastNewBatchTime)/tickPeriod) + 1
		targetTicks = utils.Min(targetTicks, ticksPerBlock) // Cap index to ticksPerBlock

		if c.ticksRunThisBlock >= targetTicks {
			// Already up to date
			return didReceiveNewBatch, false, nil
		}

		c.lock.Lock()
		defer c.lock.Unlock()

		for c.ticksRunThisBlock < targetTicks {
			c.Core.SetInBlockTickIndex(c.ticksRunThisBlock)
			c.Core.RunSingleTick()
			c.ticksRunThisBlock++
		}
		return didReceiveNewBatch, true, nil
	}
}
