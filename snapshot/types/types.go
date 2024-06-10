package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SnapshotStatus string

const (
	SnapshotStatus_Done    SnapshotStatus = "done"
	SnapshotStatus_Pending SnapshotStatus = "pending"
	SnapshotStatus_Fail    SnapshotStatus = "fail"
)

type SnapshotMetadata struct {
	Address     common.Address `json:"address"`
	BlockHash   common.Hash    `json:"blockHash"`
	BlockNumber *big.Int       `json:"blockNumber"` // TODO: make this an uint64
	StorageRoot common.Hash    `json:"storageRoot"`
}

type SnapshotMetadataWithStatus struct {
	SnapshotMetadata
	Status SnapshotStatus `json:"status"`
	Error  string         `json:"err"`
}

type SnapshotResponse struct {
	SnapshotMetadataWithStatus
	Storage []byte `json:"storage"`
}

type SnapshotQuery struct {
	BlockHash common.Hash      `json:"blockHash"`
	Addresses []common.Address `json:"addresses"`
}

type Iterator interface {
	// Next steps the iterator forward one element, returning false if exhausted,
	// or an error if iteration failed for some reason (e.g. root being iterated
	// becomes stale and garbage collected).
	Next() bool

	// Error returns any failure that occurred during iteration, which might have
	// caused a premature iteration exit (e.g. snapshot stack becoming stale).
	Error() error

	// Hash returns the hash of the account or storage slot the iterator is
	// currently at.
	Hash() common.Hash

	// Release releases associated resources. Release should always succeed and
	// can be called multiple times without causing error.
	Release()
}

type StorageIterator interface {
	Iterator

	// Slot returns the storage slot the iterator is currently at. An error will
	// be returned if the iterator becomes invalid
	Slot() []byte
}

type Schedule struct {
	Addresses   []common.Address `json:"addresses"`
	BlockPeriod uint64           `json:"blockPeriod"`
	Replace     bool             `json:"replace"`
}

type ScheduleResponse struct {
	Schedule Schedule
	ID       uint64 `json:"id"`
}
