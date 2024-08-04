package arch

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type Action interface{}

type CanonicalTickAction struct{}

// Holds all the actions included to a specific core in a specific block
type ActionBatch struct {
	BlockNumber uint64
	Actions     []Action
}

// Len returns the number of actions in the batch.
func (a ActionBatch) Len() int {
	return len(a.Actions)
}

// NewActionBatch creates a new ActionBatch instance.
func NewActionBatch(blockNumber uint64, actions []Action) ActionBatch {
	return ActionBatch{BlockNumber: blockNumber, Actions: actions}
}

type ActionBatchWithLogs struct {
	ActionBatch
	Logs []types.Log
}

func NewActionBatchWithLogs(blockNumber uint64, actions []Action, logs []types.Log) ActionBatchWithLogs {
	return ActionBatchWithLogs{ActionBatch: NewActionBatch(blockNumber, actions), Logs: logs}
}
