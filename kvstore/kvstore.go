package kvstore

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/crypto"
)

// KeyValueStore is an in-memory key-value store.
type MemoryKeyValueStore struct {
	data map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*MemoryKeyValueStore)(nil)

// NewMemoryKeyValueStore creates a new MemoryKeyValueStore.
func NewMemoryKeyValueStore() *MemoryKeyValueStore {
	return &MemoryKeyValueStore{
		data: make(map[common.Hash]common.Hash),
	}
}

// Set sets a key-value pair in the store.
func (kv *MemoryKeyValueStore) Set(key common.Hash, value common.Hash) {
	kv.data[key] = value
}

// Get gets the value for a key from the store.
func (kv *MemoryKeyValueStore) Get(key common.Hash) common.Hash {
	return kv.data[key]
}

// Has checks if a key exists in the store.
func (kv *MemoryKeyValueStore) Has(key common.Hash) bool {
	_, ok := kv.data[key]
	return ok
}

// Delete deletes a key from the store.
func (kv *MemoryKeyValueStore) Delete(key common.Hash) {
	delete(kv.data, key)
}

// Size returns the number of key-value pairs in the store.
func (kv *MemoryKeyValueStore) Size() int {
	return len(kv.data)
}

// ForEach calls a function for each key-value pair in the store.
func (kv *MemoryKeyValueStore) ForEach(forEach func(key, value common.Hash) bool) {
	for key, value := range kv.data {
		if !forEach(key, value) {
			break
		}
	}
}

// HashedMemoryKeyValueStore is an in-memory key-value store that hashes keys before storing them.
type HashedMemoryKeyValueStore struct {
	data map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*HashedMemoryKeyValueStore)(nil)

// NewHashedMemoryKeyValueStore creates a new HashedMemoryKeyValueStore.
func NewHashedMemoryKeyValueStore() *HashedMemoryKeyValueStore {
	return &HashedMemoryKeyValueStore{
		data: make(map[common.Hash]common.Hash),
	}
}

func (kv *HashedMemoryKeyValueStore) hash(key common.Hash) common.Hash {
	return crypto.Keccak256Hash(key.Bytes())
}

// Set sets a key-value pair in the store.
func (kv *HashedMemoryKeyValueStore) Set(key common.Hash, value common.Hash) {
	kv.data[kv.hash(key)] = value
}

// SetByKeyHash sets a key-value pair in the store using the key hash.
func (kv *HashedMemoryKeyValueStore) SetByKeyHash(keyHash, value common.Hash) {
	kv.data[keyHash] = value
}

// Get gets the value for a key from the store.
func (kv *HashedMemoryKeyValueStore) Get(key common.Hash) common.Hash {
	return kv.data[kv.hash(key)]
}

// GetByKeyHash gets the value for a key from the store using the key hash.
func (kv *HashedMemoryKeyValueStore) GetByKeyHash(keyHash common.Hash) common.Hash {
	return kv.data[keyHash]
}

// Has checks if a key exists in the store.
func (kv *HashedMemoryKeyValueStore) Has(key common.Hash) bool {
	_, ok := kv.data[kv.hash(key)]
	return ok
}

// HasByKeyHash checks if a key exists in the store using the key hash.
func (kv *HashedMemoryKeyValueStore) HasByKeyHash(keyHash common.Hash) bool {
	_, ok := kv.data[keyHash]
	return ok
}

// Delete deletes a key from the store.
func (kv *HashedMemoryKeyValueStore) Delete(key common.Hash) {
	delete(kv.data, kv.hash(key))
}

// DeleteByKeyHash deletes a key from the store using the key hash.
func (kv *HashedMemoryKeyValueStore) DeleteByKeyHash(keyHash common.Hash) {
	delete(kv.data, keyHash)
}

// Size returns the number of key-value pairs in the store.
func (kv *HashedMemoryKeyValueStore) Size() int {
	return len(kv.data)
}

// ForEach calls a function for each key-value pair in the store.
func (kv *HashedMemoryKeyValueStore) ForEach(forEach func(keyHash, value common.Hash) bool) {
	for key, value := range kv.data {
		if !forEach(key, value) {
			break
		}
	}
}

// CachedKeyValueStore is a key-value that all reads from the underlying store.
type CachedKeyValueStore struct {
	kv    lib.KeyValueStore
	cache map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*CachedKeyValueStore)(nil)

// NewCachedKeyValueStore creates a new CachedKeyValueStore.
func NewCachedKeyValueStore(kv lib.KeyValueStore) *CachedKeyValueStore {
	return &CachedKeyValueStore{
		kv:    kv,
		cache: make(map[common.Hash]common.Hash),
	}
}

// Set sets a key-value pair in the store.
// The write is NOT cached.
func (c *CachedKeyValueStore) Set(key, value common.Hash) {
	// Delete the cached value without caching the new value
	delete(c.cache, key)
	// Set the value in the underlying store
	c.kv.Set(key, value)
}

// Get gets the value for a key from either the cache or the underlying store.
// If the value is not in the cache, it is read from the underlying store and cached.
func (c *CachedKeyValueStore) Get(key common.Hash) common.Hash {
	if v, ok := c.cache[key]; ok {
		return v
	}
	v := c.kv.Get(key)
	c.cache[key] = v
	return v
}

// StagedKeyValueStore is a key-value store that stages writes to the underlying store until commit is called.
type StagedKeyValueStore struct {
	kv     lib.KeyValueStore
	staged map[common.Hash]common.Hash
}

var _ lib.KeyValueStore = (*StagedKeyValueStore)(nil)

// NewStagedKeyValueStore creates a new StagedKeyValueStore.
func NewStagedKeyValueStore(kv lib.KeyValueStore) *StagedKeyValueStore {
	return &StagedKeyValueStore{
		kv:     kv,
		staged: make(map[common.Hash]common.Hash),
	}
}

// Set stages a key-value pair to be written to the underlying store on commit.
func (s *StagedKeyValueStore) Set(key, value common.Hash) {
	if v := s.kv.Get(key); v == value {
		// If the value is the same as the one in the underlying store, delete the staged value
		delete(s.staged, key)
		return
	}
	s.staged[key] = value
}

// Get gets the value for a key from the store.
// If the key is staged, the staged value is returned.
func (s *StagedKeyValueStore) Get(key common.Hash) common.Hash {
	if v, ok := s.staged[key]; ok {
		return v
	}
	return s.kv.Get(key)
}

// Commit writes the staged key-value pairs to the underlying store.
func (s *StagedKeyValueStore) Commit() {
	for key, value := range s.staged {
		s.kv.Set(key, value)
	}
	s.staged = make(map[common.Hash]common.Hash)
}

// Revert discards the staged key-value pairs.
func (s *StagedKeyValueStore) Revert() {
	s.staged = make(map[common.Hash]common.Hash)
}
