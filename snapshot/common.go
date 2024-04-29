package snapshot

import (
	"time"

	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

var Root = NewSnapshotMaker(true)

type (
	SnapshotStatus             = snapshot_types.SnapshotStatus
	SnapshotMetadata           = snapshot_types.SnapshotMetadata
	SnapshotMetadataWithStatus = snapshot_types.SnapshotMetadataWithStatus
	SnapshotResponse           = snapshot_types.SnapshotResponse
	FilterQuery                = snapshot_types.FilterQuery
)

const (
	SnapshotStatus_Done    = snapshot_types.SnapshotStatus_Done
	SnapshotStatus_Pending = snapshot_types.SnapshotStatus_Pending
	SnapshotStatus_Fail    = snapshot_types.SnapshotStatus_Fail
)

const (
	SnapshotListMaxResults        = 1000
	TaskBufferSize                = 256
	SchedulerTickInterval         = 8 * time.Second
	ScheduledSnapshotsBeforePrune = 16
)

var (
	SnapshotBlobPrefix     = []byte("arch-snap-blob-")
	SnapshotMetadataPrefix = []byte("arch-snap-metadata-")
	SchedulePrefix         = []byte("arch-schedule-")
)

type Ethereum interface {
	BlockChain() *core.BlockChain
	ChainDb() ethdb.Database
}

var logger = log.Root().New("archetype", "snapshot")
