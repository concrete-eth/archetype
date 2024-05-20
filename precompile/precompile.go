package precompile

import (
	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type CorePrecompile struct {
	lib.BlankPrecompile
	schemas         arch.ArchSchemas
	coreConstructor func() arch.Core
}

var _ concrete.Precompile = (*CorePrecompile)(nil)

// NewCorePrecompile creates a new CorePrecompile.
func NewCorePrecompile(schemas arch.ArchSchemas, coreConstructor func() arch.Core) *CorePrecompile {
	return &CorePrecompile{
		schemas:         schemas,
		coreConstructor: coreConstructor,
	}
}

func (p *CorePrecompile) executeAction(env concrete.Environment, kv lib.KeyValueStore, action arch.Action) error {
	// Wrap the persistent kv store in a cached kv store to save gas when reading multiple times from the same slot
	ckv := kvstore.NewCachedKeyValueStore(kv)
	// Wrap the cached kv store in a staged kv store to save gas when writing multiple times to the same slot
	skv := kvstore.NewStagedKeyValueStore(ckv)

	// Create a new core
	core := p.coreConstructor()

	// Set the staged kv store in the core
	core.SetKV(skv)
	// Set the block number in the core
	core.SetBlockNumber(env.GetBlockNumber())

	// Execute the action
	if err := p.schemas.Actions.ExecuteAction(action, core); err != nil {
		return err
	}

	// Commit the staged kv store
	skv.Commit()

	// Emit the log
	log, err := p.schemas.Actions.ActionToLog(action)
	if err != nil {
		return err
	}
	env.Log(log.Topics, log.Data)

	return nil
}

func (p *CorePrecompile) IsStatic(input []byte) bool {
	if _, ok := p.schemas.Tables.TargetTableId(input); ok {
		return true
	}
	return false
}

func (p *CorePrecompile) Run(env concrete.Environment, input []byte) (_ret []byte, _err error) {
	// Wrap env in a kv store and datastore
	var (
		kv        = lib.NewEnvStorageKeyValueStore(env)
		datastore = lib.NewKVDatastore(kv)
	)

	// Return the data if call is a table read
	if ret, err := p.schemas.Tables.ReadPacked(datastore, input); err == nil {
		return ret, nil
	}

	// Execute the action if call is an action
	if action, err := p.schemas.Actions.CalldataToAction(input); err != nil {
		return nil, err
	} else if err := p.executeAction(env, kv, action); err != nil {
		return nil, err
	}

	return nil, nil
}
