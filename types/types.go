package types

type ActionMetadata = struct {
	Id         uint8
	Name       string
	MethodName string
}

type ActionMap = map[uint8]ActionMetadata

type Action interface{}

type CanonicalTickAction struct{}

// Holds all the actions included to a specific core in a specific block
type ActionBatch struct {
	BlockNumber uint64
	Actions     []Action
}

func NewActionBatch(blockNumber uint64, actions []Action) ActionBatch {
	return ActionBatch{BlockNumber: blockNumber, Actions: actions}
}
