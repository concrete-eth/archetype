package kvstore

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/crypto"
)

type MemoryKeyValueStore struct {
	data map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*MemoryKeyValueStore)(nil)

func NewMemoryKeyValueStore() *MemoryKeyValueStore {
	return &MemoryKeyValueStore{
		data: make(map[common.Hash]common.Hash),
	}
}

func (kv *MemoryKeyValueStore) Set(key common.Hash, value common.Hash) {
	kv.data[key] = value
}

func (kv *MemoryKeyValueStore) Get(key common.Hash) common.Hash {
	return kv.data[key]
}

func (kv *MemoryKeyValueStore) Has(key common.Hash) bool {
	_, ok := kv.data[key]
	return ok
}

func (kv *MemoryKeyValueStore) Delete(key common.Hash) {
	delete(kv.data, key)
}

func (kv *MemoryKeyValueStore) Size() int {
	return len(kv.data)
}

func (kv *MemoryKeyValueStore) ForEach(forEach func(key, value common.Hash) bool) {
	for key, value := range kv.data {
		if !forEach(key, value) {
			break
		}
	}
}

type HashedMemoryKeyValueStore struct {
	data map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*HashedMemoryKeyValueStore)(nil)

func NewHashedMemoryKeyValueStore() *HashedMemoryKeyValueStore {
	return &HashedMemoryKeyValueStore{
		data: make(map[common.Hash]common.Hash),
	}
}

func (kv *HashedMemoryKeyValueStore) hash(key common.Hash) common.Hash {
	return crypto.Keccak256Hash(key.Bytes())
}

func (kv *HashedMemoryKeyValueStore) Set(key common.Hash, value common.Hash) {
	kv.data[kv.hash(key)] = value
}

func (kv *HashedMemoryKeyValueStore) SetByKeyHash(keyHash, value common.Hash) {
	kv.data[keyHash] = value
}

func (kv *HashedMemoryKeyValueStore) Get(key common.Hash) common.Hash {
	return kv.data[kv.hash(key)]
}

func (kv *HashedMemoryKeyValueStore) GetByKeyHash(keyHash common.Hash) common.Hash {
	return kv.data[keyHash]
}

func (kv *HashedMemoryKeyValueStore) Has(key common.Hash) bool {
	_, ok := kv.data[kv.hash(key)]
	return ok
}

func (kv *HashedMemoryKeyValueStore) HasByKeyHash(keyHash common.Hash) bool {
	_, ok := kv.data[keyHash]
	return ok
}

func (kv *HashedMemoryKeyValueStore) Delete(key common.Hash) {
	delete(kv.data, kv.hash(key))
}

func (kv *HashedMemoryKeyValueStore) DeleteByKeyHash(keyHash common.Hash) {
	delete(kv.data, keyHash)
}

func (kv *HashedMemoryKeyValueStore) Size() int {
	return len(kv.data)
}

func (kv *HashedMemoryKeyValueStore) ForEach(forEach func(keyHash, value common.Hash) bool) {
	for key, value := range kv.data {
		if !forEach(key, value) {
			break
		}
	}
}

// Implements a cached key-value store that caches all reads.
type CachedKeyValueStore struct {
	kv    lib.KeyValueStore
	cache map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*CachedKeyValueStore)(nil)

func NewCachedKeyValueStore(kv lib.KeyValueStore) *CachedKeyValueStore {
	return &CachedKeyValueStore{
		kv:    kv,
		cache: make(map[common.Hash]common.Hash),
	}
}

func (c *CachedKeyValueStore) Set(key, value common.Hash) {
	// Delete the cached value without caching the new value
	delete(c.cache, key)
	// Set the value in the underlying store
	c.kv.Set(key, value)
}

func (c *CachedKeyValueStore) Get(key common.Hash) common.Hash {
	if v, ok := c.cache[key]; ok {
		return v
	}
	v := c.kv.Get(key)
	c.cache[key] = v
	return v
}

// Implements a staged key-value store that allows for atomic commits and rollbacks.
type StagedKeyValueStore struct {
	kv     lib.KeyValueStore
	staged map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*StagedKeyValueStore)(nil)

// Creates a new staged key-value store.
func NewStagedKeyValueStore(kv lib.KeyValueStore) *StagedKeyValueStore {
	return &StagedKeyValueStore{
		kv:     kv,
		staged: make(map[common.Hash]common.Hash),
	}
}

func (s *StagedKeyValueStore) Get(key common.Hash) common.Hash {
	if v, ok := s.staged[key]; ok {
		return v
	}
	return s.kv.Get(key)
}

func (s *StagedKeyValueStore) Set(key, value common.Hash) {
	if v := s.kv.Get(key); v == value {
		// If the value is the same as the one in the underlying store, delete the staged value
		delete(s.staged, key)
		return
	}
	s.staged[key] = value
}

func (s *StagedKeyValueStore) Commit() {
	for key, value := range s.staged {
		s.kv.Set(key, value)
	}
	s.staged = make(map[common.Hash]common.Hash)
}

func (s *StagedKeyValueStore) Revert() {
	s.staged = make(map[common.Hash]common.Hash)
}
