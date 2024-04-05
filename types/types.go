package types

import (
	"errors"
	"reflect"
)

type ActionMetadata = struct {
	Id         uint8
	Name       string
	MethodName string
	Type       reflect.Type
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
