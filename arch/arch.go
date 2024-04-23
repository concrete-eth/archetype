package arch

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/concrete-eth/archetype/params"
)

var (
	ErrInvalidAction   = errors.New("invalid action")
	ErrInvalidActionId = errors.New("invalid action ID")
	ErrInvalidTableId  = errors.New("invalid table ID")
)

type RawIdType = [4]byte

type validId struct {
	id    RawIdType
	valid bool
}

func (v validId) Raw() RawIdType {
	if !v.valid {
		panic("Invalid id")
	}
	return v.id
}

/*

TODO:
Schema: Spec for a single action or table
Schemas: Specs for either all actions or all tables

*/

type archSchema struct {
	datamod.TableSchema
	Method *abi.Method
	Type   reflect.Type
}

type archSchemas struct {
	abi     *abi.ABI
	schemas map[RawIdType]archSchema
}

func newArchSchemas(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
	solMethodNameFn func(string) string,
) (archSchemas, error) {
	s := archSchemas{abi: abi, schemas: make(map[RawIdType]archSchema, len(schemas))}
	for _, schema := range schemas {
		methodName := solMethodNameFn(schema.Name)
		method := abi.Methods[methodName]
		var id [4]byte
		copy(id[:], method.ID)

		actionType, ok := types[schema.Name]
		if !ok {
			return archSchemas{}, fmt.Errorf("no type found for schema %s", schema.Name)
		}

		s.schemas[id] = archSchema{
			TableSchema: datamod.TableSchema{Name: schema.Name},
			Method:      &method,
			Type:        actionType,
		}
	}
	return s, nil
}

func (a archSchemas) newId(id RawIdType) (validId, bool) {
	if _, ok := a.schemas[id]; ok {
		return validId{id: id, valid: true}, true
	}
	return validId{}, false
}

func (t archSchemas) idFromName(name string) (validId, bool) {
	for id, schema := range t.schemas {
		if schema.Name == name {
			return validId{id: id, valid: true}, true
		}
	}
	return validId{}, false
}

func (a archSchemas) getSchema(actionId validId) archSchema {
	return a.schemas[actionId.Raw()]
}

// ABI returns the ABI of the interface.
func (a archSchemas) ABI() *abi.ABI {
	return a.abi
}

type ValidActionId struct {
	validId
}

type ActionSchema struct {
	archSchema
}

type ActionSpecs struct {
	archSchemas
}

// TODO: Action schemas are table schemas without keys. Is there a better way to portray this [?]

// NewActionSpecs creates a new ActionSpecs instance.
func NewActionSpecs(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
) (ActionSpecs, error) {
	s, err := newArchSchemas(abi, schemas, types, params.SolidityActionMethodName)
	if err != nil {
		return ActionSpecs{}, err
	}
	for _, schema := range s.schemas {
		if len(schema.Method.Inputs) > 1 {
			return ActionSpecs{}, fmt.Errorf("action method %s has more than one argument", schema.Name)
		}
	}
	return ActionSpecs{archSchemas: s}, nil
}

// NewActionSpecsFromRaw creates a new ActionSpecs instance from raw JSON strings.
func NewActionSpecsFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
) (ActionSpecs, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return ActionSpecs{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return ActionSpecs{}, err
	}
	// Add the canonical Tick action
	schemas = append(schemas, datamod.TableSchema{Name: "Tick"})
	types[params.TickActionName] = reflect.TypeOf(CanonicalTickAction{})
	return NewActionSpecs(&ABI, schemas, types)
}

// NewActionId wraps a valid ID in a ValidActionId.
func (a ActionSpecs) NewActionId(id RawIdType) (ValidActionId, bool) {
	validId, ok := a.newId(id)
	return ValidActionId{validId}, ok
}

// TODO: rename
func (a ActionSpecs) ActionIdFromName(name string) (ValidActionId, bool) {
	validId, ok := a.idFromName(name)
	return ValidActionId{validId}, ok
}

// ActionIdFromAction returns the action ID of the given action.
func (a ActionSpecs) ActionIdFromAction(action Action) (ValidActionId, bool) {
	actionType := reflect.TypeOf(action)
	if !isStructPtr(actionType) {
		return ValidActionId{}, false
	}
	actionElem := actionType.Elem()
	for id, schema := range a.schemas {
		if actionElem == schema.Type {
			return ValidActionId{validId{id: id, valid: true}}, true
		}
	}
	return ValidActionId{}, false
}

// GetActionSchema returns the schema of the action with the given ID.
func (a ActionSpecs) GetActionSchema(actionId ValidActionId) ActionSchema {
	return ActionSchema{a.archSchemas.getSchema(actionId.validId)}
}

// EncodeAction encodes an action into a byte slice.
func (a *ActionSpecs) EncodeAction(action Action) (ValidActionId, []byte, error) {
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return ValidActionId{}, nil, fmt.Errorf("action of type %T does not match any canonical action type", action)
	}
	schema := a.GetActionSchema(actionId)
	data, err := packActionMethodInput(schema.Method, action)
	if err != nil {
		return ValidActionId{}, nil, err
	}
	return actionId, data, nil
}

// DecodeAction decodes the given calldata into an action.
func (a *ActionSpecs) DecodeAction(actionId ValidActionId, data []byte) (Action, error) {
	schema := a.GetActionSchema(actionId)
	args, err := schema.Method.Inputs.Unpack(data)
	if err != nil {
		return nil, err
	}
	// Create a canonically typed action from the unpacked data
	// i.e., anonymous struct{...} -> archmod.ActionData_<action name>{...}
	// All methods are autogenerated to have a single argument, so we can safely assume len(args) == 1
	action := reflect.New(schema.Type).Interface()
	if len(args) > 0 {
		if err := ConvertStruct(args[0], action); err != nil {
			return nil, err
		}
	}
	return action, nil
}

// ActionToCalldata converts an action to calldata.
// The same encoding is used for log data.
func (a *ActionSpecs) ActionToCalldata(action Action) ([]byte, error) {
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return nil, fmt.Errorf("action of type %T does not match any canonical action type", action)
	}
	schema := a.GetActionSchema(actionId)
	data, err := packActionMethodInput(schema.Method, action)
	if err != nil {
		return nil, err
	}
	calldata := make([]byte, 4+len(data))
	copy(calldata[:4], schema.Method.ID[:])
	copy(calldata[4:], data)
	return calldata, nil
}

// CalldataToAction converts calldata to an action.
func (a *ActionSpecs) CalldataToAction(calldata []byte) (Action, error) {
	if len(calldata) < 4 {
		return nil, errors.New("invalid calldata (length < 4)")
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	actionId, ok := a.NewActionId(methodId)
	if !ok {
		return nil, errors.New("method signature does not match any action")
	}
	return a.DecodeAction(actionId, calldata[4:])
}

// ActionToLog converts an action to a log.
func (a *ActionSpecs) ActionToLog(action Action) (types.Log, error) {
	data, err := a.ActionToCalldata(action)
	if err != nil {
		return types.Log{}, err
	}
	log := types.Log{
		Topics: []common.Hash{params.ActionExecutedEventID},
		Data:   data,
	}
	return log, nil
}

// LogToAction converts a log to an action.
func (a *ActionSpecs) LogToAction(log types.Log) (Action, error) {
	if len(log.Topics) != 1 || log.Topics[0] != params.ActionExecutedEventID {
		return nil, errors.New("log topics do not match action executed event")
	}
	return a.CalldataToAction(log.Data)
}

func packActionMethodInput(method *abi.Method, arg interface{}) ([]byte, error) {
	switch len(method.Inputs) {
	case 0:
		return method.Inputs.Pack()
	case 1:
		return method.Inputs.Pack(arg)
	default:
		panic("unreachable")
	}
}

type ValidTableId struct {
	validId
}

type tableGetter struct {
	constructor   reflect.Value
	rowGetterType reflect.Type
}

func newTableGetter(constructor interface{}, rowType reflect.Type) (tableGetter, error) {
	// Constructor(Datastore) -> Table
	// Table.Get(Keys) -> Row

	if constructor == nil {
		return tableGetter{}, errors.New("table constructor is nil")
	}

	constructorVal := reflect.ValueOf(constructor)
	if !constructorVal.IsValid() {
		return tableGetter{}, errors.New("table constructor is invalid")
	}
	if constructorVal.Kind() != reflect.Func {
		return tableGetter{}, errors.New("table constructor is not a function")
	}
	if constructorVal.Type().NumIn() != 1 {
		return tableGetter{}, errors.New("table constructor should take a single argument")
	}
	if constructorVal.Type().In(0) != reflect.TypeOf((*lib.Datastore)(nil)).Elem() {
		return tableGetter{}, errors.New("table constructor should take a lib.Datastore argument")
	}
	if constructorVal.Type().NumOut() != 1 {
		return tableGetter{}, errors.New("table constructor should return a single value")
	}

	tblType := constructorVal.Type().Out(0)
	getMth, ok := tblType.MethodByName("Get")
	if !ok {
		return tableGetter{}, errors.New("table missing Get method")
	}
	if getMth.Type.NumOut() != 1 {
		return tableGetter{}, errors.New("table Get method should return a single value")
	}

	retType := getMth.Type.Out(0)
	if err := CanPopulateStruct(retType, reflect.PtrTo(rowType)); err != nil {
		return tableGetter{}, fmt.Errorf("datamod row type cannot populate table row type: %v", err)
	}

	return tableGetter{
		constructor:   constructorVal,
		rowGetterType: getMth.Type,
	}, nil
}

func (t *tableGetter) get(datastore lib.Datastore, keys ...interface{}) (interface{}, error) {
	// Construct the table
	constructorArgs := []reflect.Value{reflect.ValueOf(datastore)}
	table := t.constructor.Call(constructorArgs)[0]
	// Call the Get method
	rowGetter := table.MethodByName("Get")
	// Call the Get method
	rowArgs := make([]reflect.Value, len(keys))
	for i, arg := range keys {
		argVal := reflect.ValueOf(arg)
		if argVal.Type() != t.rowGetterType.In(i) {
			return nil, fmt.Errorf("argument %d has wrong type", i)
		}
		rowArgs[i] = argVal
	}
	result := rowGetter.Call(rowArgs)[0]
	// Return the result
	return result.Interface(), nil
}

type TableSchema struct {
	archSchema
}

type TableSpecs struct {
	archSchemas
	tableGetters map[RawIdType]tableGetter
}

// NewTableSpecs creates a new TableSpecs instance.
func NewTableSpecs(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
	getters map[string]interface{},
) (TableSpecs, error) {
	s, err := newArchSchemas(abi, schemas, types, params.SolidityTableMethodName)
	if err != nil {
		return TableSpecs{}, err
	}
	for _, schema := range s.schemas {
		if len(schema.Method.Outputs) != 1 {
			return TableSpecs{}, fmt.Errorf("table method %s does not have exactly one return value", schema.Name)
		}
	}
	tableGetters := make(map[RawIdType]tableGetter, len(getters))
	for id, schema := range s.schemas {
		getterFn, ok := getters[schema.Name]
		if !ok {
			return TableSpecs{}, fmt.Errorf("no table getter found for schema %s", schema.Name)
		}
		tableGetters[id], err = newTableGetter(getterFn, schema.Type)
		if err != nil {
			return TableSpecs{}, err
		}
	}
	return TableSpecs{archSchemas: s, tableGetters: tableGetters}, nil
}

// NewTableSpecsFromRaw creates a new TableSpecs instance from raw JSON strings.
func NewTableSpecsFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
	getters map[string]interface{},
) (TableSpecs, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return TableSpecs{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return TableSpecs{}, err
	}
	return NewTableSpecs(&ABI, schemas, types, getters)
}

// NewTableId wraps a valid ID in a ValidTableId.
func (t TableSpecs) NewTableId(id RawIdType) (ValidTableId, bool) {
	validId, ok := t.newId(id)
	return ValidTableId{validId}, ok
}

func (t TableSpecs) TableIdFromName(name string) (ValidTableId, bool) {
	validId, ok := t.idFromName(name)
	return ValidTableId{validId}, ok
}

// GetTableSchema returns the schema of the table with the given ID.
func (t TableSpecs) GetTableSchema(tableId ValidTableId) TableSchema {
	return TableSchema{t.archSchemas.getSchema(tableId.validId)}
}

// read reads a row from the datastore.
func (t TableSpecs) Read(datastore lib.Datastore, tableId ValidTableId, keys ...interface{}) (interface{}, error) {
	getter := t.tableGetters[tableId.Raw()]
	dsRow, err := getter.get(datastore, keys...)
	if err != nil {
		return nil, err
	}
	schema := t.GetTableSchema(tableId)
	row := reflect.New(schema.Type).Interface()
	if err := PopulateStruct(dsRow, row); err != nil {
		return nil, err
	}
	return row, nil
}

// TableIdFromCalldata returns the table ID of the table targeted by the given calldata.
// If the calldata does not encode a table read, the second return value is false.
func (t *TableSpecs) TargetTableId(calldata []byte) (ValidTableId, bool) {
	if len(calldata) < 4 {
		return ValidTableId{}, false
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	tableId, ok := t.NewTableId(methodId)
	return tableId, ok
}

// ReadPacked reads a row from the datastore and packs it into an ABI-encoded byte slice.
func (t *TableSpecs) ReadPacked(datastore lib.Datastore, calldata []byte) ([]byte, error) {
	tableId, ok := t.TargetTableId(calldata)
	if !ok {
		return nil, errors.New("calldata does not correspond to a table read operation")
	}
	schema := t.GetTableSchema(tableId)
	keys, err := schema.Method.Inputs.UnpackValues(calldata[4:])
	if err != nil {
		return nil, err
	}
	row, err := t.Read(datastore, tableId, keys...)
	if err != nil {
		return nil, err
	}
	return schema.Method.Outputs.Pack(row)
}

type ArchSpecs struct {
	Actions ActionSpecs
	Tables  TableSpecs
}
