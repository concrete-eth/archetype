package client

import (
	"errors"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/log"

	"github.com/concrete-eth/archetype/kvstore"
	archtypes "github.com/concrete-eth/archetype/types"
)

var (
	ErrBlockNumberMismatch    = errors.New("block number mismatch")
	ErrTickNotFirst           = errors.New("tick not first action")
	ErrChannelClosed          = errors.New("channel closed")
	ErrChannelBlockedOrClosed = errors.New("channel blocked or closed")
)

type Client struct {
	Core archtypes.Core
	kv   *kvstore.StagedKeyValueStore

	actionBatchInChan <-chan archtypes.ActionBatch
	actionOutChan     chan<- []archtypes.Action

	blockTime time.Duration

	lastNewBatchTime  time.Time
	ticksRunThisBlock uint

	lock sync.Mutex
}

// New create a new client object.
func New(core archtypes.Core, kv lib.KeyValueStore, actionBatchInChan <-chan archtypes.ActionBatch, actionOutChan chan<- []archtypes.Action, blockTime time.Duration, blockNumber uint64) (*Client, error) {
	stagedKv := kvstore.NewStagedKeyValueStore(kv)
	core.SetKV(stagedKv)
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

// applyBatch Applies the given action batch to the core and returns whether a tick action was included in the batch.
// If a tick action is included, it must be the first action in the batch.
func (c *Client) applyBatch(batch archtypes.ActionBatch) (bool, error) {
	if c.Core.BlockNumber() != batch.BlockNumber {
		return false, ErrBlockNumberMismatch
	}
	tickActionInBatch := false
	for ii, action := range batch.Actions {
		if err := c.Core.ExecuteAction(action); err != nil {
			c.error("failed to execute action", "err", err)
		}
		if _, ok := action.(*archtypes.CanonicalTickAction); ok {
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
func (c *Client) applyBatchAndCommit(batch archtypes.ActionBatch) (bool, error) {
	tickActionInBatch, err := c.applyBatch(batch)
	if err != nil {
		return false, err
	}
	c.kv.Commit()
	c.lastNewBatchTime = time.Now()
	c.Core.SetBlockNumber(batch.BlockNumber + 1)
	return tickActionInBatch, nil
}

// Simulate runs the given function and then reverts all the changes to the key-value store.
func (c *Client) Simulate(f func(core archtypes.Core)) {
	// Put another stage on top of the current key-value store that will never be committed
	// and will be discarded after the function is executed
	c.lock.Lock()
	defer c.lock.Unlock()
	simKv := kvstore.NewStagedKeyValueStore(c.kv)
	c.Core.SetKV(simKv)
	f(c.Core)
	c.Core.SetKV(c.kv)
}

// SendAction is a shorthand for sending a single action to the client.
// See SendActions for more details.
func (c *Client) SendAction(action archtypes.Action) error {
	return c.SendActions([]archtypes.Action{action})
}

// SendActions sends a slice of actions to actionOutChan.
func (c *Client) SendActions(actions []archtypes.Action) error {
	actionsToSend := make([]archtypes.Action, 0)
	c.Simulate(func(core archtypes.Core) {
		for _, action := range actions {
			if err := core.ExecuteAction(action); err != nil {
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
// It will only return when the block number is reached, the channel is closed, or an error occurs.
func (c *Client) SyncUntil(blockNumber uint64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for {
		if c.Core.BlockNumber() >= blockNumber {
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
	if !c.Core.ExpectTick() {
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
