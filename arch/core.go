package arch

import (
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type Core interface {
	SetKV(kv lib.KeyValueStore) // Set the key-value store
	KV() lib.KeyValueStore      // Get the key-value store
	// ExecuteAction(Action) error // Execute the given action
	SetBlockNumber(uint64) // Set the block number
	BlockNumber() uint64   // Get the block number
	// RunSingleTick()             // Run a single tick
	// RunBlockTicks()             // Run all ticks in a block
	Tick()                      // Run a single tick
	TicksPerBlock() uint64      // Get the number of ticks per block
	ExpectTick() bool           // Check if a tick is expected
	SetInBlockTickIndex(uint64) // Set the in-block tick index
	InBlockTickIndex() uint64   // Get the in-block tick index
	Purge()                     // Purge the core
}

type ISetRebasing interface {
	SetRebasing(bool)
}

type BaseCore struct {
	kv               lib.KeyValueStore
	ds               lib.Datastore
	blockNumber      uint64
	inBlockTickIndex uint64
	rebasing         bool
}

var _ Core = &BaseCore{}

func (b *BaseCore) SetKV(kv lib.KeyValueStore) {
	b.kv = kv
	b.ds = lib.NewKVDatastore(kv)
}

func (b *BaseCore) KV() lib.KeyValueStore {
	return b.kv
}

func (b *BaseCore) Datastore() lib.Datastore {
	return b.ds
}

func (b *BaseCore) SetBlockNumber(blockNumber uint64) {
	b.blockNumber = blockNumber
}

func (b *BaseCore) BlockNumber() uint64 {
	return b.blockNumber
}

func (b *BaseCore) SetInBlockTickIndex(index uint64) {
	b.inBlockTickIndex = index
}

func (b *BaseCore) InBlockTickIndex() uint64 {
	return b.inBlockTickIndex
}

func (b *BaseCore) SetRebasing(rebasing bool) {
	b.rebasing = rebasing
}

func (b *BaseCore) Rebasing() bool {
	return b.rebasing
}

func (b *BaseCore) TicksPerBlock() uint64 {
	return 0
}

func (b *BaseCore) ExpectTick() bool {
	return true
}

func (b *BaseCore) Tick() {}

func (b *BaseCore) Purge() {}

func incrementBlockTickIndex(c Core) {
	c.SetInBlockTickIndex(c.InBlockTickIndex() + 1)
}

func RunSingleTick(c Core) {
	c.Tick()
}

func RunBlockTicks(c Core) {
	c.SetInBlockTickIndex(0)
	for i := uint64(0); i < c.TicksPerBlock(); i++ {
		RunSingleTick(c)
		incrementBlockTickIndex(c)
	}
}
