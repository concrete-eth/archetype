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
	ErrInvalidAction = errors.New("invalid action")
	// ErrInvalidActionId = errors.New("invalid action ID")
	// ErrInvalidTableId         = errors.New("invalid table ID")
	ErrCalldataIsNotTableRead = errors.New("calldata is not a table read operation")
)

type RawIdType = [4]byte

type validId struct {
	id    RawIdType
	valid bool
}

// Raw returns the raw ID.
func (v validId) Raw() RawIdType {
	if !v.valid {
		panic("Invalid id")
	}
	return v.id
}

func (v validId) Uint32() uint32 {
	return uint32(v.id[3]) | uint32(v.id[2])<<8 | uint32(v.id[1])<<16 | uint32(v.id[0])<<24
}

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
		if !isStruct(actionType) {
			return archSchemas{}, fmt.Errorf("type for schema %s is not a struct", schema.Name)
		}

		s.schemas[id] = archSchema{
			TableSchema: schema,
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

func (a archSchemas) getSchema(id validId) archSchema {
	return a.schemas[id.Raw()]
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

type ActionSchemas struct {
	archSchemas
}

// NewActionSchemas creates a new ActionSchemas instance.
func NewActionSchemas(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
) (ActionSchemas, error) {
	if _, ok := types[params.TickActionName]; !ok {
		// Add the canonical Tick action
		schemas = append(schemas, datamod.TableSchema{
			Name:   "Tick",
			Keys:   []datamod.FieldSchema{},
			Values: []datamod.FieldSchema{},
		})
		types[params.TickActionName] = reflect.TypeOf(CanonicalTickAction{})
	}
	s, err := newArchSchemas(abi, schemas, types, params.SolidityActionMethodName)
	if err != nil {
		return ActionSchemas{}, err
	}
	for _, schema := range s.schemas {
		if len(schema.Method.Inputs) > 1 {
			return ActionSchemas{}, fmt.Errorf("action method %s has more than one argument", schema.Name)
		}
	}
	return ActionSchemas{archSchemas: s}, nil
}

// NewActionSchemasFromRaw creates a new ActionSchemas instance from raw JSON strings.
func NewActionSchemasFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
) (ActionSchemas, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return ActionSchemas{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return ActionSchemas{}, err
	}
	return NewActionSchemas(&ABI, schemas, types)
}

// NewActionId wraps a valid ID in a ValidActionId.
func (a ActionSchemas) NewActionId(id RawIdType) (ValidActionId, bool) {
	validId, ok := a.newId(id)
	return ValidActionId{validId}, ok
}

// ActionIdFromName returns the action ID of the action with the given name.
func (a ActionSchemas) ActionIdFromName(name string) (ValidActionId, bool) {
	validId, ok := a.idFromName(name)
	return ValidActionId{validId}, ok
}

// ActionIdFromAction returns the action ID of the given action.
func (a ActionSchemas) ActionIdFromAction(action Action) (ValidActionId, bool) {
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
func (a ActionSchemas) GetActionSchema(actionId ValidActionId) ActionSchema {
	return ActionSchema{a.archSchemas.getSchema(actionId.validId)}
}

// EncodeAction encodes an action into a byte slice.
func (a *ActionSchemas) EncodeAction(action Action) (ValidActionId, []byte, error) {
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
func (a *ActionSchemas) DecodeAction(actionId ValidActionId, data []byte) (Action, error) {
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
		if err := ConvertStruct(action, args[0]); err != nil {
			return nil, err
		}
	}
	return action, nil
}

// ActionToCalldata converts an action to calldata.
// The same encoding is used for log data.
func (a *ActionSchemas) ActionToCalldata(action Action) ([]byte, error) {
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
func (a *ActionSchemas) CalldataToAction(calldata []byte) (Action, error) {
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
func (a *ActionSchemas) ActionToLog(action Action) (types.Log, error) {
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
func (a *ActionSchemas) LogToAction(log types.Log) (Action, error) {
	if len(log.Topics) != 1 || log.Topics[0] != params.ActionExecutedEventID {
		return nil, errors.New("log topics do not match ActionExecuted event")
	}
	return a.CalldataToAction(log.Data)
}

// ExecuteAction executes the given action on the given target.
func (a *ActionSchemas) ExecuteAction(action Action, target Core) error {
	if _, ok := action.(*CanonicalTickAction); ok {
		RunBlockTicks(target)
		return nil
	}
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return ErrInvalidAction
	}
	schema := a.GetActionSchema(actionId)
	actionName := schema.Name
	methodName := actionName
	targetVal := reflect.ValueOf(target)
	if !targetVal.IsValid() {
		return fmt.Errorf("target is invalid")
	}
	method := targetVal.MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method %s not found", methodName)
	}
	args := []reflect.Value{reflect.ValueOf(action)}
	result := method.Call(args)
	if len(result) == 0 {
		return nil
	}
	errVal := result[len(result)-1]
	if !errVal.IsNil() {
		return errVal.Interface().(error)
	}
	return nil
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
		return tableGetter{}, errors.New("table constructor does not have exactly one argument")
	}
	if constructorVal.Type().In(0) != reflect.TypeOf((*lib.Datastore)(nil)).Elem() {
		return tableGetter{}, errors.New("table constructor must take a lib.Datastore argument")
	}
	if constructorVal.Type().NumOut() != 1 {
		return tableGetter{}, errors.New("table constructor must return a single value")
	}

	tblType := constructorVal.Type().Out(0)
	getMth, ok := tblType.MethodByName("Get")
	if !ok {
		return tableGetter{}, errors.New("table missing Get method")
	}
	if getMth.Type.NumOut() != 1 {
		return tableGetter{}, errors.New("table Get method must return a single value")
	}

	retType := getMth.Type.Out(0)
	if err := CanPopulateStruct(reflect.PtrTo(rowType), retType); err != nil {
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
		if argVal.Type() != t.rowGetterType.In(i+1) { // First argument is the receiver
			return nil, fmt.Errorf("key for argument %d has wrong type: expected %v, got %v", i, t.rowGetterType.In(i+1), argVal.Type())
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

type TableSchemas struct {
	archSchemas
	tableGetters map[RawIdType]tableGetter
}

// NewTableSchemas creates a new TableSchemas instance.
func NewTableSchemas(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
	getters map[string]interface{},
) (TableSchemas, error) {
	s, err := newArchSchemas(abi, schemas, types, params.SolidityTableMethodName)
	if err != nil {
		return TableSchemas{}, err
	}
	for _, schema := range s.schemas {
		if len(schema.Method.Outputs) != 1 {
			return TableSchemas{}, fmt.Errorf("table method %s does not have exactly one return value", schema.Name)
		}
	}
	tableGetters := make(map[RawIdType]tableGetter, len(getters))
	for id, schema := range s.schemas {
		getterFn, ok := getters[schema.Name]
		if !ok {
			return TableSchemas{}, fmt.Errorf("no table getter found for schema %s", schema.Name)
		}
		tableGetters[id], err = newTableGetter(getterFn, schema.Type)
		if err != nil {
			return TableSchemas{}, err
		}
	}
	return TableSchemas{archSchemas: s, tableGetters: tableGetters}, nil
}

// NewTableSchemasFromRaw creates a new TableSchemas instance from raw JSON strings.
func NewTableSchemasFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
	getters map[string]interface{},
) (TableSchemas, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return TableSchemas{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return TableSchemas{}, err
	}
	return NewTableSchemas(&ABI, schemas, types, getters)
}

// NewTableId wraps a valid ID in a ValidTableId.
func (t TableSchemas) NewTableId(id RawIdType) (ValidTableId, bool) {
	validId, ok := t.newId(id)
	return ValidTableId{validId}, ok
}

// TableIdFromName returns the table ID for the given name, if valid.
func (t TableSchemas) TableIdFromName(name string) (ValidTableId, bool) {
	validId, ok := t.idFromName(name)
	return ValidTableId{validId}, ok
}

// GetTableSchema returns the schema of the table with the given ID.
func (t TableSchemas) GetTableSchema(tableId ValidTableId) TableSchema {
	return TableSchema{t.archSchemas.getSchema(tableId.validId)}
}

// read reads a row from the datastore.
func (t TableSchemas) Read(datastore lib.Datastore, tableId ValidTableId, keys ...interface{}) (interface{}, error) {
	getter := t.tableGetters[tableId.Raw()]
	dsRow, err := getter.get(datastore, keys...)
	if err != nil {
		return nil, err
	}
	schema := t.GetTableSchema(tableId)
	row := reflect.New(schema.Type).Interface()
	if err := PopulateStruct(row, dsRow); err != nil {
		return nil, err
	}
	return row, nil
}

// TableIdFromCalldata returns the table ID of the table targeted by the given calldata.
// If the calldata does not encode a table read, the second return value is false.
func (t *TableSchemas) TargetTableId(calldata []byte) (ValidTableId, bool) {
	if len(calldata) < 4 {
		return ValidTableId{}, false
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	tableId, ok := t.NewTableId(methodId)
	return tableId, ok
}

// ReadPacked reads a row from the datastore and packs it into an ABI-encoded byte slice.
func (t *TableSchemas) ReadPacked(datastore lib.Datastore, calldata []byte) ([]byte, error) {
	tableId, ok := t.TargetTableId(calldata)
	if !ok {
		return nil, ErrCalldataIsNotTableRead
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

type ArchSchemas struct {
	Actions ActionSchemas
	Tables  TableSchemas
}
