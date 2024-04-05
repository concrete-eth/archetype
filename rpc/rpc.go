package rpc

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/concrete-eth/archetype/params"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

/*

- Subscribe to action batches: must be cancelable
- Send action batches in transactions

*/

var (
	StandardTimeout        = 5 * time.Second // Standard timeout for RPC requests
	BlockQueryLimit uint64 = 256             // Maximum number of blocks to query in a single request
	HeaderChanSize         = 4               // Size of the header channel
)

type ActionBatchSubscription struct {
	ethcli          EthCli
	actionMap       archtypes.ActionMap
	actionsAbi      abi.ABI
	coreAddress     common.Address
	actionBatchChan chan<- archtypes.ActionBatch
	unsubChan       chan struct{}
	errChan         chan error
	closeUnsubOnce  sync.Once
	closeErrOnce    sync.Once
	unsubscribed    bool
}

// SubscribeActionBatches subscribes to action batches emitted by the core contract at coreAddress.
func SubscribeActionBatches(
	ethcli EthCli,
	actionMap archtypes.ActionMap,
	actionsAbi abi.ABI,
	coreAddress common.Address,
	startingBlockNumber uint64,
	actionBatchesChan chan<- archtypes.ActionBatch,
) *ActionBatchSubscription {
	sub := &ActionBatchSubscription{
		ethcli:          ethcli,
		actionMap:       actionMap,
		actionsAbi:      actionsAbi,
		coreAddress:     coreAddress,
		actionBatchChan: actionBatchesChan,
		unsubChan:       make(chan struct{}),
		errChan:         make(chan error, 1),
	}
	go sub.runSubscription(startingBlockNumber)
	return sub
}

func (s *ActionBatchSubscription) tryCloseUnsub() {
	s.closeUnsubOnce.Do(func() {
		close(s.unsubChan)
	})
}

func (s *ActionBatchSubscription) tryCloseErr() {
	s.closeErrOnce.Do(func() {
		close(s.errChan)
	})
}

func (s *ActionBatchSubscription) hasUnsubscribed() bool {
	if s.unsubscribed {
		return true
	}
	select {
	case <-s.unsubChan:
		s.unsubscribed = true
		s.tryCloseUnsub()
		s.tryCloseErr()
		return true
	default:
		return false
	}
}

func (s *ActionBatchSubscription) sendErr(err error) {
	s.errChan <- err
	s.tryCloseUnsub()
	s.tryCloseErr()
}

func (s *ActionBatchSubscription) runSubscription(startingBlock uint64) {
	if _, err := s.sync(startingBlock); err != nil {
		s.sendErr(err)
		return
	}
}

func (s *ActionBatchSubscription) getHeadBlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	return s.ethcli.BlockNumber(ctx)
}

func (s *ActionBatchSubscription) getLogs(fromBlock, toBlock uint64) ([]types.Log, error) {
	ctx, cancel := context.WithTimeout(context.Background(), StandardTimeout)
	defer cancel()
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{s.coreAddress},
		Topics:    [][]common.Hash{{s.actionsAbi.Events[params.ActionExecutedEventName].ID}},
	}
	return s.ethcli.FilterLogs(ctx, query)
}

// sync sends an action batch for every block from startingBlock to the head block.
// When the head block is reached, a new batch is sent for every new block.
// sync will only return when the subscription is unsubscribed or an error occurs.
func (s *ActionBatchSubscription) sync(startingBlock uint64) (uint64, error) {
	var oldestUnsyncedBN uint64
	var err error
	if oldestUnsyncedBN, err = s.syncToHead(startingBlock); err != nil {
		return oldestUnsyncedBN, err
	}
	if oldestUnsyncedBN, err = s.syncAtHead(oldestUnsyncedBN); err != nil {
		return oldestUnsyncedBN, err
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) syncToHead(startingBlock uint64) (uint64, error) {
	if s.hasUnsubscribed() {
		return startingBlock, nil
	}
	oldestUnsyncedBN := startingBlock
	headBN, err := s.getHeadBlockNumber()
	if err != nil {
		return startingBlock, err
	}

	for oldestUnsyncedBN < headBN {
		if s.hasUnsubscribed() {
			return oldestUnsyncedBN, nil
		}
		var fromBN, toBN uint64
		{
			// Sync from oldestUnsyncedBN to oldestUnsyncedBN + BlockQueryLimit
			// If toBN > headBN, refetch headBN and set toBN = min(headBN, toBN)
			fromBN = oldestUnsyncedBN
			toBN = oldestUnsyncedBN + BlockQueryLimit
			if toBN > headBN {
				headBN, err := s.getHeadBlockNumber()
				if err != nil {
					return oldestUnsyncedBN, err
				}
				toBN = utils.Min(headBN, toBN)
			}
		}
		// Fetch logs from fromBN to toBN
		logs, err := s.getLogs(fromBN, toBN)
		if err != nil {
			return oldestUnsyncedBN, err
		}
		// Process logs
		if oldestUnsyncedBN, err = s.processLogs(logs, oldestUnsyncedBN, toBN); err != nil {
			return oldestUnsyncedBN, err
		}
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) syncAtHead(startingBlock uint64) (uint64, error) {
	if s.hasUnsubscribed() {
		return startingBlock, nil
	}
	oldestUnsyncedBN := startingBlock

	headerChan := make(chan *types.Header, HeaderChanSize)
	headersSub, err := s.ethcli.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		return oldestUnsyncedBN, err
	}
	defer headersSub.Unsubscribe()

	for {
		select {
		case err := <-headersSub.Err():
			return oldestUnsyncedBN, err
		case header := <-headerChan:
			if s.hasUnsubscribed() {
				return oldestUnsyncedBN, nil
			}
			if header.Number.Uint64() < oldestUnsyncedBN {
				continue
			}
			// Fetch logs from oldestUnsyncedBN to head
			logs, err := s.getLogs(oldestUnsyncedBN, header.Number.Uint64())
			if err != nil {
				return oldestUnsyncedBN, err
			}
			// Process logs
			if oldestUnsyncedBN, err = s.processLogs(logs, oldestUnsyncedBN, header.Number.Uint64()); err != nil {
				return oldestUnsyncedBN, err
			}
		}
	}
}

func (s *ActionBatchSubscription) processLogs(logs []types.Log, from, to uint64) (uint64, error) {
	oldestUnsyncedBN := from
	logBatch := make([]types.Log, 0)
	for _, log := range logs {
		for oldestUnsyncedBN < log.BlockNumber {
			if s.hasUnsubscribed() {
				return oldestUnsyncedBN, nil
			}
			if err := s.sendLogBatch(oldestUnsyncedBN, logBatch); err != nil {
				return oldestUnsyncedBN, err
			}
			logBatch = make([]types.Log, 0)
			oldestUnsyncedBN++
		}
		logBatch = append(logBatch, log)
	}
	for oldestUnsyncedBN <= to {
		if err := s.sendLogBatch(oldestUnsyncedBN, logBatch); err != nil {
			return oldestUnsyncedBN, err
		}
		logBatch = make([]types.Log, 0)
		oldestUnsyncedBN++
	}
	return oldestUnsyncedBN, nil
}

func (s *ActionBatchSubscription) sendLogBatch(blockNumber uint64, logBatch []types.Log) error {
	// Process logBatch into action batch and send
	actions := make([]archtypes.Action, 0, len(logBatch))
	for _, log := range logBatch {
		action, err := s.logToAction(log)
		if err != nil {
			return err
		}
		actions = append(actions, action)
	}
	actionBatch := archtypes.NewActionBatch(blockNumber, actions)
	s.actionBatchChan <- actionBatch
	return nil
}

func (s *ActionBatchSubscription) logToAction(log types.Log) (archtypes.Action, error) {
	event := s.actionsAbi.Events[params.ActionExecutedEventName]

	// Check if log is an ActionExecuted event
	if len(log.Topics) < 1 {
		return nil, errors.New("no topics in log")
	}
	if log.Topics[0] != event.ID {
		return nil, errors.New("not an ActionExecuted event")
	}

	// Unpack log data
	args, err := event.Inputs.Unpack(log.Data)
	if err != nil {
		return nil, err
	}

	// Get action ID
	actionId := args[0].(uint8)
	actionMetadata, ok := s.actionMap[actionId]
	if !ok {
		return nil, errors.New("unknown action ID")
	}

	// Get action data
	method := s.actionsAbi.Methods[actionMetadata.MethodName]
	var anonAction interface{}
	if len(method.Inputs) == 0 {
		anonAction = struct{}{}
	} else {
		_actionData := args[1].([]byte)
		_action, err := method.Inputs.Unpack(_actionData)
		if err != nil {
			return nil, err
		}
		anonAction = _action[0]
	}

	// TODO: anon to cannon

	return anonAction, nil
}

func (s *ActionBatchSubscription) Unsubscribe() {
	s.unsubChan <- struct{}{}
}

func (s *ActionBatchSubscription) Err() <-chan error {
	return s.errChan
}

var _ ethereum.Subscription = (*ActionBatchSubscription)(nil)
