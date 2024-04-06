package precompile

import (
	"fmt"

	archcodec "github.com/concrete-eth/archetype/codec"
	"github.com/concrete-eth/archetype/kvstore"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/lib"
	ccutils "github.com/ethereum/go-ethereum/concrete/utils"
)

type CorePrecompile struct {
	lib.BlankPrecompile
	coreAbi   abi.ABI
	getDataFn func(lib.Datastore, *abi.Method, []interface{}) (interface{}, bool)
	core      archtypes.Core
}

var _ concrete.Precompile = (*CorePrecompile)(nil)

func (p *CorePrecompile) unpackInput(input []byte) (method *abi.Method, args []interface{}, err error) {
	methodId, data := ccutils.SplitInput(input)
	method, err = p.coreAbi.MethodById(methodId)
	if err != nil {
		return
	}
	args, err = method.Inputs.Unpack(data)
	if err != nil {
		return nil, nil, err
	}
	return method, args, nil
}

func (p *CorePrecompile) runMethod(env concrete.Environment, method *abi.Method, args []interface{}) (interface{}, error) {
	var (
		kv        = lib.NewEnvPersistentKeyValueStore(env)
		datastore = lib.NewKVDatastore(kv)
	)
	if data, ok := p.getDataFn(datastore, method, args); ok {
		return data, nil
	}
	action, err := archcodec.ArgsToAction(method, args)
	if err != nil {
		return nil, err
	}
	if err := p.executeAction(env, kv, method, 0, action); err != nil { // TODO
		return nil, err
	}
	return nil, nil
}

func (p *CorePrecompile) executeAction(env concrete.Environment, kv lib.KeyValueStore, method *abi.Method, actionId uint8, action archtypes.Action) error {
	// Wrap the persistent kv store in a cached kv store to save gas when reading multiple times from the same slot
	ckv := kvstore.NewCachedKeyValueStore(kv)
	// Wrap the cached kv store in a staged kv store to save gas when writing multiple times to the same slot
	skv := kvstore.NewStagedKeyValueStore(ckv)
	// Set the staged kv store in the core
	p.core.SetKV(skv)
	// Set the block number in the core
	p.core.SetBlockNumber(env.GetBlockNumber())
	// Execute the action
	if err := p.core.ExecuteAction(action); err != nil {
		return err
	}
	// Commit the staged kv store
	skv.Commit()
	// Emit the log
	log, err := archcodec.ActionToLogWithMethod(method, actionId, action)
	if err != nil {
		return err
	}
	env.Log(log.Topics, log.Data)

	return nil
}

func (p *CorePrecompile) IsStatic(input []byte) bool {
	methodID, _ := ccutils.SplitInput(input)
	method, err := p.coreAbi.MethodById(methodID)
	if err != nil {
		return false
	}
	return method.IsConstant()
}

func (p *CorePrecompile) Run(env concrete.Environment, input []byte) (_ret []byte, _err error) {
	defer func() {
		if r := recover(); r != nil {
			_ret = nil
			_err = fmt.Errorf("panic: %v", r)
		}
	}()

	method, args, err := p.unpackInput(input)
	if err != nil {
		return nil, err
	}

	result, err := p.runMethod(env, method, args)
	if err != nil {
		return nil, err
	}

	if len(method.Outputs) == 0 {
		_ret, err = method.Outputs.Pack()
	} else if len(method.Outputs) == 1 {
		_ret, err = method.Outputs.Pack(result)
	} else {
		_ret, err = method.Outputs.Pack(result.([]interface{})...)
	}
	if err != nil {
		return nil, err
	}

	return _ret, nil
}
