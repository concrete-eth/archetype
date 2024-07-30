package client

import (
	"errors"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/log"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/kvstore"
)

var (
	ErrBlockNumberMismatch    = errors.New("block number mismatch")
	ErrTickNotFirst           = errors.New("tick not first action")
	ErrChannelClosed          = errors.New("channel closed")
	ErrChannelBlockedOrClosed = errors.New("channel blocked or closed")
)

type Client struct {
	schemas arch.ArchSchemas

	core arch.Core
	kv   *kvstore.StagedKeyValueStore

	actionBatchInChan <-chan arch.ActionBatch
	actionOutChan     chan<- []arch.Action

	blockTime time.Duration

	lastNewBatchTime  time.Time
	ticksRunThisBlock uint64

	lock sync.Mutex

	now func() time.Time

	_tickTime time.Duration
}

// New create a new client object.
func New(schemas arch.ArchSchemas, core arch.Core, kv lib.KeyValueStore, actionBatchChan <-chan arch.ActionBatch, actionChan chan<- []arch.Action, blockTime time.Duration, blockNumber uint64) *Client {
	stagedKv := kvstore.NewStagedKeyValueStore(kv)
	core.SetKV(stagedKv)
	return &Client{
		schemas:           schemas,
		core:              core,
		kv:                stagedKv,
		actionBatchInChan: actionBatchChan,
		actionOutChan:     actionChan,
		blockTime:         blockTime,
		ticksRunThisBlock: 0,
		now:               time.Now,

		_tickTime: blockTime / time.Duration(core.TicksPerBlock()),
	}
}

// Core returns the core.
func (c *Client) Core() arch.Core {
	return c.core
}

// BlockTime returns the block time.
func (c *Client) BlockTime() time.Duration {
	return c.blockTime
}

func (c *Client) TickTime() time.Duration {
	return c._tickTime
}

func (c *Client) LastNewBatchTime() time.Time {
	return c.lastNewBatchTime
}

// func (c *Client) debug(msg string, ctx ...interface{}) {
// 	log.Debug(msg, ctx...)
// }

// func (c *Client) warn(msg string, ctx ...interface{}) {
// 	log.Warn(msg, ctx...)
// }

func (c *Client) error(msg string, ctx ...interface{}) {
	log.Error(msg, ctx...)
}

// applyBatch Applies the given action batch to the core and returns whether a tick action was included in the batch.
// If a tick action is included, it must be the first action in the batch.
func (c *Client) applyBatch(batch arch.ActionBatch) (bool, error) {
	if c.core.BlockNumber() != batch.BlockNumber {
		return false, ErrBlockNumberMismatch
	}
	tickActionInBatch := false
	for ii, action := range batch.Actions {
		if err := c.schemas.Actions.ExecuteAction(action, c.core); err != nil {
			c.error("failed to execute action", "err", err)
		}
		if _, ok := action.(*arch.CanonicalTickAction); ok {
			if ii != 0 {
				return false, ErrTickNotFirst
			}
			tickActionInBatch = true
		}
	}
	return tickActionInBatch, nil
}

// applyBatchAndCommit applies the given action batch to the core, commits the changes to the key-value store,
// and updates the core block number.
func (c *Client) applyBatchAndCommit(batch arch.ActionBatch) (bool, error) {
	tickActionInBatch, err := c.applyBatch(batch)
	if err != nil {
		return false, err
	}
	c.kv.Commit()
	c.lastNewBatchTime = c.now()
	c.core.SetBlockNumber(batch.BlockNumber + 1)
	return tickActionInBatch, nil
}

// Simulate runs the given function and then reverts all the changes to the key-value store.
func (c *Client) Simulate(f func(core arch.Core)) {
	// Put another stage on top of the current key-value store that will never be committed
	// and will be discarded after the function is executed
	c.lock.Lock()
	defer c.lock.Unlock()
	simKv := kvstore.NewStagedKeyValueStore(c.kv)
	c.core.SetKV(simKv)
	f(c.core)
	c.core.SetKV(c.kv)
}

// SendAction is a shorthand for sending a single action to the client.
// See SendActions for more details.
func (c *Client) SendAction(action arch.Action) error {
	return c.SendActions([]arch.Action{action})
}

// SendActions sends a slice of actions to actionOutChan.
func (c *Client) SendActions(actions []arch.Action) error {
	actionsToSend := make([]arch.Action, 0)
	c.Simulate(func(core arch.Core) {
		for _, action := range actions {
			if err := c.schemas.Actions.ExecuteAction(action, core); err != nil {
				c.error("failed to execute action", "err", err)
				continue
			}
			actionsToSend = append(actionsToSend, action)
		}
	})

	select {
	case c.actionOutChan <- actionsToSend:
		return nil
	default:
		return ErrChannelBlockedOrClosed
	}
}

// Sync apply all buffered action batches and commit the changes to the key-value store.
// Returns whether a new batch of actions was received and whether a tick action was executed.
// If no batches are available, it will do nothing and return false.
func (c *Client) Sync() (didReceiveNewBatch bool, didTick bool, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	select {
	case batch, ok := <-c.actionBatchInChan:
		if !ok {
			return false, false, ErrChannelClosed
		}
		didTick, err := c.applyBatchAndCommit(batch)
		return true, didTick, err
	default:
		return false, false, nil
	}
}

// SyncUntil applies action batches until the block number is reached.
// The action batch corresponding to the given block number will not be included.
// It will only return when the block number is reached, the channel is closed, or an error occurs.
func (c *Client) SyncUntil(blockNumber uint64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for {
		if c.core.BlockNumber() >= blockNumber {
			break
		}
		batch, ok := <-c.actionBatchInChan
		if !ok {
			return ErrChannelClosed
		}
		if _, err := c.applyBatchAndCommit(batch); err != nil {
			return err
		}
	}
	return nil
}

// InterpolatedSync applies an action batch if available, otherwise it anticipates the ticks expected in the next block.
// When an action batch is received, it will revert any tick anticipation and apply the batch normally.
func (c *Client) InterpolatedSync() (didReceiveNewBatch bool, didTick bool, err error) {
	if !c.core.ExpectTick() {
		return c.Sync()
	}
	select {
	case batch, ok := <-c.actionBatchInChan:
		if !ok {
			return false, false, ErrChannelClosed
		}
		// Received a new batch of actions
		// Revert any tick anticipation and apply batch normally
		didReceiveNewBatch = true
		c.kv.Revert()
		c.core.SetInBlockTickIndex(0)

		if base, ok := c.core.(arch.ISetRebasing); ok {
			base.SetRebasing(true)
		}
		didTick, err = c.applyBatchAndCommit(batch)
		if err != nil {
			return didReceiveNewBatch, didTick, err
		}
		if base, ok := c.core.(arch.ISetRebasing); ok {
			base.SetRebasing(false)
		}

		c.ticksRunThisBlock = 0
		c.core.SetInBlockTickIndex(0)
	default:
		// No new batch of actions received
		// Anticipate ticks corresponding to the expected tick action in the next block
		didReceiveNewBatch = false
	}

	var (
		ticksPerBlock = c.core.TicksPerBlock()
		tickTime      = c.TickTime()
	)

	targetTicks := uint64(c.now().Sub(c.lastNewBatchTime)/tickTime) + 1
	targetTicks = utils.Min(targetTicks, ticksPerBlock) // Cap index to ticksPerBlock

	if c.ticksRunThisBlock >= targetTicks {
		// Already up to date
		return didReceiveNewBatch, didTick, nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	for c.ticksRunThisBlock < targetTicks {
		c.core.SetInBlockTickIndex(c.ticksRunThisBlock)
		arch.RunSingleTick(c.core)
		c.ticksRunThisBlock++
	}
	return didReceiveNewBatch, true, nil
}
