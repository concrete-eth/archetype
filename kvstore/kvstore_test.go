package kvstore

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

func testKeyValueStore(t *testing.T, kv lib.KeyValueStore) {
	key := common.BytesToHash([]byte("foobar"))
	val0 := common.BytesToHash([]byte("foo"))
	val1 := common.BytesToHash([]byte("bar"))
	kv.Set(key, val0)
	if kv.Get(key) != val0 {
		t.Errorf("expected %v, got %v", val0, kv.Get(key))
	}
	kv.Set(key, val1)
	if kv.Get(key) != val1 {
		t.Errorf("expected %v, got %v", val1, kv.Get(key))
	}
}

func TestMemoryKeyValueStore(t *testing.T) {
	testKeyValueStore(t, NewMemoryKeyValueStore())
}

func TestHashedMemoryKeyValueStore(t *testing.T) {
	testKeyValueStore(t, NewHashedMemoryKeyValueStore())

	kv := NewHashedMemoryKeyValueStore()

	var (
		key     = common.BytesToHash([]byte("foobar"))
		keyHash = crypto.Keccak256Hash(key.Bytes())
		val0    = common.BytesToHash([]byte("foo"))
		val1    = common.BytesToHash([]byte("bar"))
	)

	kv.Set(key, val0)
	if kv.GetByKeyHash(keyHash) != val0 {
		t.Errorf("expected %v, got %v", val0, kv.GetByKeyHash(keyHash))
	}

	kv.SetByKeyHash(keyHash, val1)
	if kv.Get(key) != val1 {
		t.Errorf("expected %v, got %v", val1, kv.Get(key))
	}

	if !kv.Has(key) {
		t.Errorf("expected key %v to exist", key)
	}
	if !kv.HasByKeyHash(keyHash) {
		t.Errorf("expected key hash %v to exist", keyHash)
	}

	kv.Delete(key)

	if kv.Get(key) != (common.Hash{}) {
		t.Errorf("expected %v, got %v", common.Hash{}, kv.Get(key))
	}
	if kv.GetByKeyHash(keyHash) != (common.Hash{}) {
		t.Errorf("expected %v, got %v", common.Hash{}, kv.GetByKeyHash(keyHash))
	}
	if kv.Has(key) {
		t.Errorf("expected key %v to not exist", key)
	}
	if kv.HasByKeyHash(keyHash) {
		t.Errorf("expected key hash %v to not exist", keyHash)
	}
}

func TestCachedKeyValueStore(t *testing.T) {
	testKeyValueStore(t, NewCachedKeyValueStore(NewMemoryKeyValueStore()))

	baseKv := &hookedKeyValueStore{KeyValueStore: NewMemoryKeyValueStore()}
	cachedKv := NewCachedKeyValueStore(baseKv)

	key := common.BytesToHash([]byte("foo"))
	val := common.BytesToHash([]byte("bar"))

	baseKv.Set(key, val)
	if cachedKv.Get(key) != val {
		t.Errorf("expected %v, got %v", val, cachedKv.Get(key))
	}

	baseKv.getHook = func(key common.Hash) {
		t.Errorf("Unexpected call to baseKv.Get(%v)", key)
	}

	if cachedKv.Get(key) != val {
		t.Errorf("expected %v, got %v", val, cachedKv.Get(key))
	}
}

func TestStagedKeyValueStore(t *testing.T) {
	testKeyValueStore(t, NewStagedKeyValueStore(NewMemoryKeyValueStore()))

	baseKv := NewMemoryKeyValueStore()
	stagedKv := NewStagedKeyValueStore(baseKv)

	var (
		key  = common.BytesToHash([]byte("foobar"))
		val0 = common.BytesToHash([]byte("foo"))
		val1 = common.BytesToHash([]byte("bar"))
	)

	stagedKv.Set(key, val0)
	if baseKv.Get(key) != (common.Hash{}) {
		t.Errorf("expected %v, got %v", common.Hash{}, baseKv.Get(key))
	}

	stagedKv.Set(key, val1)
	stagedKv.Commit()
	if baseKv.Get(key) != val1 {
		t.Errorf("expected %v, got %v", val1, baseKv.Get(key))
	}
}

type hookedKeyValueStore struct {
	lib.KeyValueStore
	getHook func(common.Hash)
	setHook func(common.Hash, common.Hash)
}

func (kv *hookedKeyValueStore) Get(key common.Hash) common.Hash {
	if kv.getHook != nil {
		kv.getHook(key)
	}
	return kv.KeyValueStore.Get(key)
}

func (kv *hookedKeyValueStore) Set(key common.Hash, value common.Hash) {
	if kv.setHook != nil {
		kv.setHook(key, value)
	}
	kv.KeyValueStore.Set(key, value)
}
