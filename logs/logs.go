package logs

import (
	"errors"
	"reflect"

	"github.com/concrete-eth/archetype/params"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func LogToAction(actionAbi abi.ABI, actionMap archtypes.ActionMap, log types.Log) (interface{}, error) {
	event := actionAbi.Events[params.ActionExecutedEventName]

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
	actionMetadata, ok := actionMap[actionId]
	if !ok {
		return nil, errors.New("unknown action ID")
	}

	// Get action data
	method := actionAbi.Methods[actionMetadata.MethodName]
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

	// Create action
	action := reflect.New(actionMetadata.Type)

	// Copy action data to action
	if err := archtypes.ConvertStruct(anonAction, action); err != nil {
		return nil, err
	}

	return action, nil
}

func ActionToLog(actionAbi abi.ABI, actionMap archtypes.ActionMap, actionId uint8, action archtypes.Action) (types.Log, error) {
	actionMetadata, ok := actionMap[actionId]
	if !ok {
		return types.Log{}, errors.New("unknown action ID")
	}

	method := actionAbi.Methods[actionMetadata.MethodName]
	_actionData, err := method.Inputs.Pack(action)
	if err != nil {
		return types.Log{}, err
	}

	event := actionAbi.Events[params.ActionExecutedEventName]
	data, err := event.Inputs.PackValues([]interface{}{actionId, _actionData})
	if err != nil {
		return types.Log{}, err
	}

	return types.Log{
		Topics: []common.Hash{event.ID},
		Data:   data,
	}, nil
}
