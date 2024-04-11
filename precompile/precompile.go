package precompile

import (
	"fmt"

	"github.com/concrete-eth/archetype/kvstore"
	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type CorePrecompile struct {
	lib.BlankPrecompile
	spec archtypes.ArchSpecs
	core archtypes.Core
}

var _ concrete.Precompile = (*CorePrecompile)(nil)

// NewCorePrecompile creates a new CorePrecompile.
func NewCorePrecompile(spec archtypes.ArchSpecs, core archtypes.Core) *CorePrecompile {
	return &CorePrecompile{
		spec: spec,
		core: core,
	}
}

func (p *CorePrecompile) executeAction(env concrete.Environment, kv lib.KeyValueStore, action archtypes.Action) error {
	// Wrap the persistent kv store in a cached kv store to save gas when reading multiple times from the same slot
	ckv := kvstore.NewCachedKeyValueStore(kv)
	// Wrap the cached kv store in a staged kv store to save gas when writing multiple times to the same slot
	skv := kvstore.NewStagedKeyValueStore(ckv)

	// Set the staged kv store in the core
	p.core.SetKV(skv)
	// Set the block number in the core
	p.core.SetBlockNumber(env.GetBlockNumber())

	// Execute the action
	if err := archtypes.ExecuteAction(p.spec.Actions, action, p.core); err != nil {
		return err
	}

	// Commit the staged kv store
	skv.Commit()

	// Emit the log
	log, err := p.spec.Actions.ActionToLog(action)
	if err != nil {
		return err
	}
	env.Log(log.Topics, log.Data)

	return nil
}

func (p *CorePrecompile) IsStatic(input []byte) bool {
	if _, ok := p.spec.Tables.TargetTableId(input); ok {
		return true
	}
	return false
}

func (p *CorePrecompile) Run(env concrete.Environment, input []byte) (_ret []byte, _err error) {
	defer func() {
		if r := recover(); r != nil {
			_ret = nil
			_err = fmt.Errorf("panic: %v", r)
		}
	}()

	// Wrap env in a kv store and datastore
	var (
		kv        = lib.NewEnvPersistentKeyValueStore(env)
		datastore = lib.NewKVDatastore(kv)
	)

	// Return the data if call is a table read
	if ret, err := p.spec.Tables.ReadPacked(datastore, input); err == nil {
		return ret, nil
	}

	// Execute the action if call is an action
	action, err := p.spec.Actions.CalldataToAction(input)
	if err != nil {
		return nil, err
	}
	if err := p.executeAction(env, kv, action); err != nil {
		return nil, err
	}

	return nil, nil
}
