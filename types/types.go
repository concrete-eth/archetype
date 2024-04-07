package types

import (
	"errors"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type ActionMetadata = struct {
	Id         uint8
	Name       string
	MethodName string
	Type       reflect.Type
}

type ActionMap = map[uint8]ActionMetadata

type ActionSpecs struct {
	Actions ActionMap
	ABI     *abi.ABI
}

func (a ActionSpecs) ActionIdFromAction(action interface{}) (uint8, bool) {
	actionType := reflect.TypeOf(action)
	for id, metadata := range a.Actions {
		if metadata.Type == actionType {
			return id, true
		}
	}
	return 0, false
}

type TableMetadata = struct {
	Id         uint8
	Name       string
	MethodName string
	Keys       []string
	Columns    []string
	Type       reflect.Type
}

type TableMap = map[uint8]TableMetadata

type TableSpecs struct {
	Tables  map[uint8]TableMetadata
	ABI     *abi.ABI
	GetData func(datastore lib.Datastore, method *abi.Method, args []interface{}) (interface{}, bool)
}

type ArchSpecs struct {
	Actions ActionSpecs
	Tables  TableSpecs
}

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

// ConvertStruct copies the fields from src to dest if they have the same name and type.
func ConvertStruct(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Struct {
		return errors.New("src is not a struct")
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Struct {
		return errors.New("dest is not a pointer to a struct")
	}

	destElem := destVal.Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		destField := destElem.FieldByName(srcVal.Type().Field(i).Name)

		if destField.IsValid() && destField.CanSet() && srcField.Type() == destField.Type() {
			destField.Set(srcField)
		}
	}

	return nil
}

type Core interface {
	SetKV(kv lib.KeyValueStore) error // Set the key-value store
	ExecuteAction(Action) error       // Execute the given action
	SetBlockNumber(uint64)            // Set the block number
	BlockNumber() uint64              // Get the block number
	RunSingleTick()                   // Run a single tick
	RunBlockTicks()                   // Run all ticks in a block
	TicksPerBlock() uint              // Get the number of ticks per block
	ExpectTick() bool                 // Check if a tick is expected
	SetInBlockTickIndex(uint)         // Set the in-block tick index
	InBlockTickIndex() uint           // Get the in-block tick index
}
