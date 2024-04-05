package types

import "math/big"

type BlockNumber struct {
	value uint64
}

func NewBlockNumber(value uint64) BlockNumber {
	return BlockNumber{value: value}
}

func (b BlockNumber) Uint64() uint64 { return uint64(b.value) }

func (b BlockNumber) Big() *big.Int { return new(big.Int).SetUint64(uint64(b.value)) }

func (b BlockNumber) AddUint64(value uint64) BlockNumber {
	return BlockNumber{value: b.value + value}
}

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
	BlockNumber BlockNumber
	Actions     []Action
}

func NewActionBatch(blockNumber BlockNumber, actions []Action) ActionBatch {
	return ActionBatch{BlockNumber: blockNumber, Actions: actions}
}
