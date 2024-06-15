package snapshot

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/concrete-eth/archetype/snapshot/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// TODO: disambiguate naming collision between contract storage blob snapshots and geth snapshots
// TODO: docstrings

var (
	ErrBlockNotFound     = errors.New("block not found")
	ErrSnapshotNotFound  = errors.New("snapshot not found")
	ErrSchedulerDisabled = errors.New("scheduler is disabled")
	ErrMissingBlob       = errors.New("missing blob")
)

type SnapshotWriter interface {
	New(query SnapshotQuery) ([]SnapshotMetadataWithStatus, error)
	Update(query SnapshotQuery) ([]SnapshotMetadataWithStatus, error)
	Delete(query SnapshotQuery) error
	Prune() error
	AddSchedule(schedule snapshot_types.Schedule) (snapshot_types.ScheduleResponse, error)
	DeleteSchedule(id uint64) error
}

type SnapshotReader interface {
	Get(address common.Address, blockHash common.Hash) (SnapshotResponse, error)
	Last(address common.Address) (SnapshotMetadataWithStatus, error)
	List(address common.Address) ([]SnapshotMetadataWithStatus, error)
	GetSchedules() (map[uint64]snapshot_types.Schedule, error)
}

type WorkerTask struct {
	Metadata  SnapshotMetadata
	Replace   bool
	BlockRoot common.Hash
	BlockHash common.Hash
	CopyFrom  common.Address
}

type SnapshotMaker struct {
	readWriter *snapshotReaderWriter
}

func NewSnapshotMaker(enableScheduling bool) *SnapshotMaker {
	return &SnapshotMaker{
		readWriter: &snapshotReaderWriter{
			taskQueueChan:    make(chan *WorkerTask, TaskBufferSize),
			schedulerEnabled: enableScheduling,
			snapshotsPending: make(map[common.Hash]common.Address),
			snapshotsFailed:  make(map[common.Hash]error),
		},
	}
}

func initReadWriter(rw *snapshotReaderWriter, ethereum Ethereum) {
	blockNumber := ethereum.BlockChain().CurrentHeader().Number.Uint64()
	rw.eth = ethereum
	rw.db = ethereum.ChainDb()
	rw.triedb = state.NewDatabase(rw.db)
	if rw.IsSchedulerEnabled() {
		rw.scheduler = NewScheduler(rw.db, blockNumber)
	}
}

func (s *SnapshotMaker) NewWriter(ethereum Ethereum) SnapshotWriter {
	if s.readWriter.eth == nil {
		initReadWriter(s.readWriter, ethereum)
	}
	return s.readWriter
}

func (s *SnapshotMaker) NewReader(ethereum Ethereum) SnapshotReader {
	if s.readWriter.eth == nil {
		initReadWriter(s.readWriter, ethereum)
	}
	return s.readWriter
}

func (s *SnapshotMaker) IsSchedulerEnabled() bool {
	return s.readWriter.IsSchedulerEnabled()
}

func (s *SnapshotMaker) RunSnapshotWorker() {
	s.readWriter.RunSnapshotWorker()
}

func (s *SnapshotMaker) RunScheduler() {
	s.readWriter.RunScheduler()
}

type snapshotReaderWriter struct {
	eth    Ethereum
	db     ethdb.Database
	triedb state.Database

	taskQueueChan chan *WorkerTask

	schedulerEnabled             bool
	scheduler                    *Scheduler
	schedulerCreationsSincePrune uint
	snapshotsPending             map[common.Hash]common.Address
	snapshotsFailed              map[common.Hash]error

	lock       sync.RWMutex
	prunerLock sync.Mutex
}

func (s *snapshotReaderWriter) IsSchedulerEnabled() bool {
	return s.schedulerEnabled
}

func (s *snapshotReaderWriter) RunSnapshotWorker() {
	for task := range s.taskQueueChan {
		s.runSnapshotWorkerTask(task)
	}
}

// setPending marks a snapshot as pending
// address is the address of the first account (not necessarily the only account) for which a snapshot with that storage root is pending
func (s *snapshotReaderWriter) setPending(addr common.Address, storageRoot common.Hash) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// A root will only be removed from failed when it is marked as pending
	// as that means it is queued for re-generation
	delete(s.snapshotsFailed, storageRoot)
	if _, ok := s.snapshotsPending[storageRoot]; ok {
		// Already pending
		return
	}
	s.snapshotsPending[storageRoot] = addr
}

func (s *snapshotReaderWriter) setDone(storageRoot common.Hash) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.snapshotsPending, storageRoot)
}

func (s *snapshotReaderWriter) getPending(storageRoot common.Hash) (common.Address, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	addr, ok := s.snapshotsPending[storageRoot]
	return addr, ok
}

func (s *snapshotReaderWriter) setFailed(storageRoot common.Hash, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.snapshotsFailed[storageRoot] = err
}

func (s *snapshotReaderWriter) getFailed(storageRoot common.Hash) (bool, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	err, ok := s.snapshotsFailed[storageRoot]
	return ok, err
}

func (s *snapshotReaderWriter) runSnapshotWorkerTask(task *WorkerTask) {
	// Roots will only be marked as failed from this function

	metadata := task.Metadata

	var blobZip []byte

	if task.CopyFrom != (common.Address{}) {
		// Copy blob from another address
		if blob := ReadSnapshotBlob(s.db, task.CopyFrom, metadata.StorageRoot); blob != nil {
			blobZip = blob
		}
	}
	if blobZip == nil {
		log.Info("Running snapshot blob generation", "address", task.Metadata.Address, "storageRoot", task.Metadata.StorageRoot)
		var err error
		blobZip, err = s.makeBlob(metadata.Address, task.BlockHash, task.BlockRoot)
		if err != nil {
			log.Error("Failed to generate snapshot blob", "address", metadata.Address, "storageRoot", metadata.StorageRoot, "err", err)
			s.setFailed(metadata.StorageRoot, fmt.Errorf("failed to generate blob: %w", err))
			return
		}
		log.Info("Finished snapshot blob generation", "address", metadata.Address, "storageRoot", metadata.StorageRoot)
	}

	// Wait for pruner to finish before writing new blobs
	s.prunerLock.Lock()
	defer s.prunerLock.Unlock()

	WriteSnapshotBlob(s.db, metadata.Address, metadata.StorageRoot, blobZip)
	s.setDone(metadata.StorageRoot)

	if task.Replace {
		oldestMetadata, _ := s.getOldestAndNewest(metadata.Address)
		if oldestMetadata.BlockHash != (common.Hash{}) {
			DeleteSnapshotMetadata(s.db, oldestMetadata.Address, oldestMetadata.BlockHash)
			// Blob is not deleted as it is may be referenced by intermediate metadata
		}
	}

	log.Info("Finished (over) writing snapshot", "address", metadata.Address, "storageRoot", metadata.StorageRoot)
}

func (s *snapshotReaderWriter) makeBlob(address common.Address, blockHash, blockRoot common.Hash) ([]byte, error) {
	targetBlockHeader := s.eth.BlockChain().GetHeaderByHash(blockHash)
	targetBlockNumber := targetBlockHeader.Number.Uint64()

	_, previousBlobMetadata := s.getOldestAndNewestInRange(address, 0, targetBlockNumber)
	if previousBlobMetadata.BlockHash == (common.Hash{}) {
		// No previous blob to build on top of
		return s.makeBlobFromScratch(address, blockRoot)
	}

	nLayers := int(targetBlockNumber - previousBlobMetadata.BlockNumber.Uint64())
	layers := s.eth.BlockChain().Snapshots().Snapshots(blockRoot, nLayers, false)

	storage := make(map[common.Hash][]byte)
	addressHash := crypto.Keccak256Hash(address.Bytes())

	for _, layer := range layers {
		it, destructed := layer.(interface {
			StorageIterator(account common.Hash, seek common.Hash) (snapshot.StorageIterator, bool)
		}).StorageIterator(addressHash, common.Hash{})
		for it.Next() {
			slot := it.Hash()
			if _, ok := storage[slot]; ok {
				// Do not override with past value
				continue
			}
			value := it.Slot()
			storage[slot] = value
		}
		if destructed {
			break
		}
	}

	prevBlob := ReadSnapshotBlob(s.db, address, previousBlobMetadata.StorageRoot)
	rawBlob, err := utils.Decompress(prevBlob)
	if err != nil {
		return nil, err
	}

	it := utils.BlobToStorageIt(rawBlob)
	for it.Next() {
		slot := it.Hash()
		if _, ok := storage[slot]; ok {
			// Do not override with past value
			continue
		}
		value := it.Slot()
		storage[slot] = value
	}

	var blob, blobZip []byte
	if blob, err = utils.MappingToBlob(storage); err != nil {
		return nil, err
	}
	if blobZip, err = utils.Compress(blob); err != nil {
		return nil, err
	}

	return blobZip, nil
}

// TODO: revise logging

func (s *snapshotReaderWriter) makeBlobFromScratch(address common.Address, blockRoot common.Hash) ([]byte, error) {
	addressHash := crypto.Keccak256Hash(address.Bytes())
	storageLeafSnapshotIt, err := s.eth.BlockChain().Snapshots().StorageIterator(blockRoot, addressHash, common.Hash{})
	if err != nil {
		return nil, err
	}
	var blob, blobZip []byte
	if blob, err = utils.StorageItToBlob(storageLeafSnapshotIt); err != nil {
		return nil, err
	}
	if blobZip, err = utils.Compress(blob); err != nil {
		return nil, err
	}
	return blobZip, nil
}

func (s *snapshotReaderWriter) RunScheduler() {
	ticker := time.NewTicker(SchedulerTickInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.runScheduler()
	}
}

func (s *snapshotReaderWriter) runScheduler() {
	// Prune dangling snapshots if more than 16 snapshots have been created by the scheduler since last prune
	// Prune before running the scheduler to give time for snapshot creations from the previous scheduler run to finish
	if s.schedulerCreationsSincePrune >= ScheduledSnapshotsBeforePrune {
		s.Prune()
		s.schedulerCreationsSincePrune = 0
	}
	var (
		currentHeader = s.eth.BlockChain().CurrentHeader()
		blockNumber   = currentHeader.Number.Uint64()
		blockHash     = currentHeader.Hash()
	)
	s.scheduler.RunSchedule(blockNumber, func(schedule snapshot_types.Schedule) {
		var res []SnapshotMetadataWithStatus
		var err error
		query := SnapshotQuery{
			Addresses: schedule.Addresses,
			BlockHash: blockHash,
		}
		if schedule.Replace {
			res, err = s.Update(query)
		} else {
			res, err = s.New(query)
		}
		if err != nil {
			log.Error("Scheduled snapshot generation failed", "err", err)
			return
		}
		for _, metadata := range res {
			// Assume none of the pending snapshot were in progress beforehand and, therefore, all pending snapshots
			// were just queued during this scheduler run
			if metadata.Status == SnapshotStatus_Pending {
				s.schedulerCreationsSincePrune++
			}
		}
	})
}

func (s *snapshotReaderWriter) snapshotStatus(metadata SnapshotMetadata) (SnapshotStatus, error) {
	if _, pending := s.getPending(metadata.StorageRoot); pending {
		return SnapshotStatus_Pending, nil
	} else if ok, err := s.getFailed(metadata.StorageRoot); ok {
		return SnapshotStatus_Fail, err
	} else if HasSnapshotBlob(s.db, metadata.Address, metadata.StorageRoot) {
		return SnapshotStatus_Done, nil
	} else if blob := ReadSnapshotBlobAnyAccount(s.db, metadata.StorageRoot); blob != nil {
		// Blob is already in database from another address with same storage root
		return SnapshotStatus_Done, nil
	} else {
		log.Error("Missing snapshot blob", "address", metadata.Address, "storageRoot", metadata.StorageRoot)
		return SnapshotStatus_Fail, ErrMissingBlob
	}
}

func (s *snapshotReaderWriter) New(query SnapshotQuery) (r []SnapshotMetadataWithStatus, err error) {
	return s.newOrUpdate(query, false)
}

func (s *snapshotReaderWriter) Update(query SnapshotQuery) (r []SnapshotMetadataWithStatus, err error) {
	return s.newOrUpdate(query, true)
}

func (s *snapshotReaderWriter) newOrUpdate(query SnapshotQuery, replace bool) (r []SnapshotMetadataWithStatus, err error) {
	addresses := query.Addresses
	blockHash := query.BlockHash
	defer func() {
		if err == nil {
			log.Info("New snapshots", "blockHash", blockHash, "addresses", len(addresses))
		} else {
			log.Error("Failed to create new snapshots", "err", err, "blockHash", blockHash, "addresses", len(addresses))
		}
	}()
	header := s.eth.BlockChain().GetHeaderByHash(blockHash)
	if header == nil {
		return nil, ErrBlockNotFound
	}
	blockNumber := header.Number
	blockRoot := header.Root

	found, missing := s.lookupSnapshots(addresses, blockHash)

	var snapshot snapshot.Snapshot
	if snapshots := s.eth.BlockChain().Snapshots(); snapshots != nil {
		snapshot = snapshots.Snapshot(blockRoot)
	}
	trie, err := s.triedb.OpenTrie(blockRoot)
	if err != nil {
		return nil, err
	}

	response := make([]SnapshotMetadataWithStatus, 0, len(found)+len(missing))
	for address := range found {
		storageRoot := s.getStorageRoot(snapshot, trie, address)
		metadata := SnapshotMetadata{
			Address:     address,
			BlockHash:   blockHash,
			BlockNumber: blockNumber,
			StorageRoot: storageRoot,
		}
		status, err := s.snapshotStatus(metadata)
		response = append(response, SnapshotMetadataWithStatus{
			SnapshotMetadata: metadata,
			Status:           status,
			Error:            errToString(err),
		})
	}
	if len(missing) == 0 {
		return response, nil
	}

	batch := s.db.NewBatch()
	for address := range missing {
		storageRoot := s.getStorageRoot(snapshot, trie, address)
		metadata := SnapshotMetadata{
			Address:     address,
			BlockHash:   blockHash,
			BlockNumber: blockNumber,
			StorageRoot: storageRoot,
		}
		metadataWithStatus := SnapshotMetadataWithStatus{
			SnapshotMetadata: metadata,
		}
		WriteSnapshotMetadata(batch, metadata)
		if copyFrom, ok := s.getPending(storageRoot); ok {
			if err := s.queueSnapshotReplication(metadata, replace, blockRoot, copyFrom); err == nil {
				metadataWithStatus.Status = SnapshotStatus_Pending
			} else {
				metadataWithStatus.Status = SnapshotStatus_Fail
				metadataWithStatus.Error = errToString(err)
				// s.setFailed(storageRoot, err)
			}
		} else {
			if HasSnapshotBlob(s.db, address, storageRoot) {
				// Blob is already in database from another block with same storage root
				metadataWithStatus.Status = SnapshotStatus_Done
			} else if blob := ReadSnapshotBlobAnyAccount(s.db, storageRoot); blob != nil {
				// Blob is already in database from another address with same storage root
				// Copy it under the current address
				WriteSnapshotBlob(batch, address, storageRoot, blob)
				metadataWithStatus.Status = SnapshotStatus_Done
			} else {
				s.setPending(address, storageRoot)
				// Blob is not in database, queue snapshot generation
				if err := s.queueSnapshotGeneration(metadata, replace, blockRoot); err == nil {
					metadataWithStatus.Status = SnapshotStatus_Pending
				} else {
					metadataWithStatus.Status = SnapshotStatus_Fail
					metadataWithStatus.Error = errToString(err)
					s.setFailed(storageRoot, err)
				}
			}
		}
		response = append(response, metadataWithStatus)

		if replace && metadataWithStatus.Status == SnapshotStatus_Done {
			// Delete old snapshot metadata if any
			// Blob is not deleted as it is referenced by the new metadata
			oldestMetadata, _ := s.getOldestAndNewest(metadata.Address)
			if oldestMetadata.BlockHash != (common.Hash{}) {
				DeleteSnapshotMetadata(batch, oldestMetadata.Address, oldestMetadata.BlockHash)
			}
		}
	}
	if err := batch.Write(); err != nil {
		return nil, err
	}
	return response, nil
}

func (s *snapshotReaderWriter) queueTask(task *WorkerTask) error {
	select {
	case s.taskQueueChan <- task:
		log.Info("Queued snapshot generation", "address", task.Metadata.Address, "storageRoot", task.Metadata.StorageRoot)
		return nil
	default:
		log.Warn("Task queue is full, skipping snapshot generation", "address", task.Metadata.Address, "storageRoot", task.Metadata.StorageRoot)
		return errors.New("failed to enqueue snapshot generation: task queue full")
	}
}

func (s *snapshotReaderWriter) queueSnapshotGeneration(metadata SnapshotMetadata, replace bool, blockRoot common.Hash) error {
	task := &WorkerTask{
		Metadata:  metadata,
		Replace:   replace,
		BlockRoot: blockRoot,
		BlockHash: metadata.BlockHash,
	}
	return s.queueTask(task)
}

func (s *snapshotReaderWriter) queueSnapshotReplication(metadata SnapshotMetadata, replace bool, blockRoot common.Hash, copyFrom common.Address) error {
	task := &WorkerTask{
		Metadata:  metadata,
		Replace:   replace,
		BlockRoot: blockRoot,
		BlockHash: metadata.BlockHash,
		CopyFrom:  copyFrom,
	}
	return s.queueTask(task)
}

func (s *snapshotReaderWriter) Delete(query SnapshotQuery) (err error) {
	defer func() {
		if err == nil {
			log.Info("Deleted snapshots", "blockHash", query.BlockHash, "addresses", len(query.Addresses))
		} else {
			log.Error("Failed to delete snapshots", "err", err, "blockHash", query.BlockHash, "addresses", len(query.Addresses))
		}
	}()
	if len(query.Addresses) > 0 {
		if query.BlockHash != (common.Hash{}) {
			return s.deleteByAddressAndBlock(query)
		} else {
			return s.deleteByAddressAllBlocks(query.Addresses)
		}
	}
	if query.BlockHash != (common.Hash{}) {
		return s.deleteAllAddressesByBlock(query.BlockHash)
	}
	return s.deleteAll()
}

func (s *snapshotReaderWriter) deleteByAddressAndBlock(query SnapshotQuery) error {
	found, _ := s.lookupSnapshots(query.Addresses, query.BlockHash)
	batch := s.db.NewBatch()
	for address := range found {
		// Only the metadata is deleted as the blob might be shared with other block hash
		DeleteSnapshotMetadata(batch, address, query.BlockHash)
	}
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *snapshotReaderWriter) deleteByAddressAllBlocks(addresses []common.Address) error {
	db := s.eth.ChainDb()
	batch := db.NewBatch()
	for _, address := range addresses {
		it := IterateAccountSnapshotMetadata(db, address)
		for it.Next() {
			metadata := it.Metadata()
			// Both metadata and blob are deleted as snapshots are deleted across all blocks
			DeleteSnapshotMetadata(batch, address, metadata.BlockHash)
			DeleteSnapshotBlob(batch, address, metadata.StorageRoot)
		}
		it.Release()
	}
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *snapshotReaderWriter) deleteAllAddressesByBlock(blockHash common.Hash) error {
	found := s.lookupAllSnapshots(blockHash)
	batch := s.db.NewBatch()
	for address := range found {
		// Only the metadata is deleted as the blob might be shared with other block hash
		DeleteSnapshotMetadata(batch, address, blockHash)
	}
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *snapshotReaderWriter) deleteAll() error {
	db := s.eth.ChainDb()
	batch := db.NewBatch()
	it := IterateSnapshotMetadata(db)
	for it.Next() {
		address := it.Address()
		metadata := it.Metadata()
		// Both metadata and blob are deleted as snapshots are deleted across all blocks
		DeleteSnapshotMetadata(batch, address, metadata.BlockHash)
		DeleteSnapshotBlob(batch, address, metadata.StorageRoot)
	}
	it.Release()
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *snapshotReaderWriter) Prune() (err error) {
	defer func() {
		if err == nil {
			log.Info("Pruned dangling snapshots")
		} else {
			log.Error("Failed to prune dangling snapshots", "err", err)
		}
	}()

	// Lock pruner to prevent new blobs from being written between the time we
	// iterate over metadata and over blobs
	s.prunerLock.Lock()
	defer s.prunerLock.Unlock()

	// Find all storage roots that are referenced by metadata
	used := make(map[common.Hash]struct{})
	metaIt := IterateSnapshotMetadata(s.db)
	for metaIt.Next() {
		metadata := metaIt.Metadata()
		storageRoot := metadata.StorageRoot
		used[storageRoot] = struct{}{}
	}
	metaIt.Release()
	// Delete all blobs that are not referenced by metadata
	batch := s.db.NewBatch()
	blobIt := IterateSnapshotBlobs(s.db)
	for blobIt.Next() {
		storageRoot := blobIt.StorageRoot()
		if _, ok := used[storageRoot]; !ok {
			batch.Delete(blobIt.Key())
		}
	}
	blobIt.Release()
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (s *snapshotReaderWriter) Get(address common.Address, blockHash common.Hash) (r SnapshotResponse, err error) {
	metadata, found := s.lookupSnapshot(address, blockHash)
	if !found {
		return SnapshotResponse{}, ErrSnapshotNotFound
	}

	response := SnapshotResponse{
		SnapshotMetadataWithStatus: SnapshotMetadataWithStatus{
			SnapshotMetadata: metadata,
		},
	}

	status, err := s.snapshotStatus(metadata)
	response.Status = status
	response.Error = errToString(err)

	if status != SnapshotStatus_Done {
		return response, nil
	}

	blob := ReadSnapshotBlob(s.db, address, metadata.StorageRoot)
	if blob == nil {
		log.Error("Missing snapshot blob", "address", address, "storageRoot", metadata.StorageRoot)
		response.Status = SnapshotStatus_Fail
		response.Error = ErrMissingBlob.Error()
	}
	response.Storage = blob

	return response, nil
}

func (s *snapshotReaderWriter) Last(address common.Address) (r SnapshotMetadataWithStatus, err error) {
	_, newestMetadata := s.getOldestAndNewest(address)
	if newestMetadata.BlockHash == (common.Hash{}) {
		return SnapshotMetadataWithStatus{}, ErrSnapshotNotFound
	}
	return SnapshotMetadataWithStatus{
		SnapshotMetadata: newestMetadata,
		Status:           SnapshotStatus_Done,
	}, nil
}

func (s *snapshotReaderWriter) getOldestAndNewest(address common.Address) (oldest, newest SnapshotMetadata) {
	return s.getOldestAndNewestInRange(address, 0, math.MaxUint64)
}

func (s *snapshotReaderWriter) getOldestAndNewestInRange(address common.Address, l, r uint64) (oldest, newest SnapshotMetadata) {
	var minMetadata, maxMetadata SnapshotMetadata
	it := IterateAccountSnapshotMetadata(s.db, address)
	defer it.Release()
	for it.Next() {
		blockHash := it.BlockHash()
		status, _ := s.snapshotStatus(it.Metadata())
		if status != SnapshotStatus_Done {
			// Skip snapshots that are not done
			continue
		}
		header := s.eth.BlockChain().GetHeaderByHash(blockHash)
		if header == nil {
			continue
		}
		blockNumber := header.Number.Uint64()
		if blockNumber < l || blockNumber > r {
			continue
		}
		canonicalHeader := s.eth.BlockChain().GetHeaderByNumber(blockNumber)
		if canonicalHeader == nil || canonicalHeader.Hash() != blockHash {
			// Skip non-canonical blocks
			continue
		}
		if minMetadata.BlockHash == (common.Hash{}) || blockNumber < minMetadata.BlockNumber.Uint64() {
			minMetadata = it.Metadata()
		}
		if maxMetadata.BlockHash == (common.Hash{}) || blockNumber > maxMetadata.BlockNumber.Uint64() {
			maxMetadata = it.Metadata()
		}
	}
	return minMetadata, maxMetadata
}

func (s *snapshotReaderWriter) List(address common.Address) (r []SnapshotMetadataWithStatus, err error) {
	db := s.eth.ChainDb()
	it := IterateAccountSnapshotMetadata(db, address)
	defer it.Release()
	response := make([]SnapshotMetadataWithStatus, 0)
	for it.Next() {
		if len(response) >= SnapshotListMaxResults {
			break
		}
		metadata := it.Metadata()
		metadataWithStatus := SnapshotMetadataWithStatus{
			SnapshotMetadata: metadata,
		}
		status, err := s.snapshotStatus(metadata)
		if err == ErrMissingBlob {
			continue
		}
		metadataWithStatus.Status = status
		metadataWithStatus.Error = errToString(err)
		response = append(response, metadataWithStatus)
	}
	return response, nil
}

func (s *snapshotReaderWriter) AddSchedule(schedule snapshot_types.Schedule) (snapshot_types.ScheduleResponse, error) {
	if !s.IsSchedulerEnabled() {
		return snapshot_types.ScheduleResponse{}, ErrSchedulerDisabled
	}
	blockNumber := s.eth.BlockChain().CurrentHeader().Number.Uint64()
	return s.scheduler.AddSchedule(schedule, blockNumber)
}

func (s *snapshotReaderWriter) DeleteSchedule(id uint64) error {
	if !s.IsSchedulerEnabled() {
		return ErrSchedulerDisabled
	}
	return s.scheduler.DeleteSchedule(id)
}

func (s *snapshotReaderWriter) GetSchedules() (map[uint64]snapshot_types.Schedule, error) {
	if !s.IsSchedulerEnabled() {
		return nil, ErrSchedulerDisabled
	}
	return s.scheduler.GetSchedules()
}

// Return the storage root for the given address
// Snapshot, if not nil, must correspond to the same block root as trie
func (s *snapshotReaderWriter) getStorageRoot(snapshot snapshot.Snapshot, trie state.Trie, address common.Address) (storageRoot common.Hash) {
	if snapshot != nil {
		// Lookup snapshot if available
		addressHash := crypto.Keccak256Hash(address.Bytes())
		account, err := snapshot.Account(addressHash)
		if err == nil {
			if account == nil {
				// Account does not exist
				return types.EmptyRootHash
			}
			storageRoot = utils.DecodeSnapshotStorageRoot(account.Root)
			return storageRoot
		}
	}
	// Lookup trie
	account, err := trie.GetAccount(address)
	if err != nil {
		// Trie is corrupted
		log.Crit("Failed to get account", "err", err)
		return common.Hash{}
	}
	if account == nil {
		// Account does not exist
		return types.EmptyRootHash
	}
	storageRoot = account.Root
	return storageRoot
}

func (s *snapshotReaderWriter) lookupSnapshot(address common.Address, blockHash common.Hash) (metadata SnapshotMetadata, found bool) {
	metadata = ReadSnapshotMetadata(s.db, address, blockHash)
	if metadata.BlockHash == (common.Hash{}) {
		return SnapshotMetadata{}, false
	}
	return metadata, true
}

// Finds snapshots for which metadata is available in the DB.
func (s *snapshotReaderWriter) lookupSnapshots(addresses []common.Address, blockHash common.Hash) (found, missing map[common.Address]struct{}) {
	found = make(map[common.Address]struct{})
	missing = make(map[common.Address]struct{})
	db := s.eth.ChainDb()
	for _, address := range addresses {
		if HasSnapshotMetadata(db, address, blockHash) {
			found[address] = struct{}{}
		} else {
			missing[address] = struct{}{}
		}
	}
	return found, missing
}

// Finds snapshots for the given block hash
func (s *snapshotReaderWriter) lookupAllSnapshots(blockHash common.Hash) (found map[common.Address]struct{}) {
	found = make(map[common.Address]struct{})
	it := IterateSnapshotMetadata(s.db)
	defer it.Release()
	for it.Next() {
		address := it.Address()
		foundBlockHash := it.BlockHash()
		if foundBlockHash == blockHash {
			found[address] = struct{}{}
		}
	}
	return found
}

func rootSnapshotBlobKeyPrefix(storageRoot common.Hash) []byte {
	return append(SnapshotBlobPrefix, storageRoot.Bytes()...)
}

func snapshotBlobKey(storageRoot common.Hash, address common.Address) []byte {
	return append(rootSnapshotBlobKeyPrefix(storageRoot), address.Bytes()...)
}

func accountSnapshotMetadataKeyPrefix(address common.Address) []byte {
	return append(SnapshotMetadataPrefix, address.Bytes()...)
}

func snapshotMetadataKey(address common.Address, blockHash common.Hash) []byte {
	return append(accountSnapshotMetadataKeyPrefix(address), blockHash.Bytes()...)
}

func WriteSnapshotBlob(db ethdb.KeyValueWriter, address common.Address, storageRoot common.Hash, storage []byte) {
	if err := db.Put(snapshotBlobKey(storageRoot, address), storage); err != nil {
		log.Crit("Failed to write snapshot", "err", err)
	}
}

func DeleteSnapshotBlob(db ethdb.KeyValueWriter, address common.Address, storageRoot common.Hash) {
	if err := db.Delete(snapshotBlobKey(storageRoot, address)); err != nil {
		log.Crit("Failed to delete snapshot", "err", err)
	}
}

func ReadSnapshotBlob(db ethdb.KeyValueReader, address common.Address, storageRoot common.Hash) []byte {
	enc, err := db.Get(snapshotBlobKey(storageRoot, address))
	if err != nil {
		return nil
	}
	return enc
}

func HasSnapshotBlob(db ethdb.KeyValueReader, address common.Address, storageRoot common.Hash) bool {
	ok, _ := db.Has(snapshotBlobKey(storageRoot, address))
	return ok
}

func ReadSnapshotBlobAnyAccount(db ethdb.Iteratee, storageRoot common.Hash) []byte {
	it := db.NewIterator(rootSnapshotBlobKeyPrefix(storageRoot), nil)
	defer it.Release()
	if !it.Next() {
		return nil
	}
	return it.Value()
}

func HasSnapshotBlobAnyAccount(db ethdb.Iteratee, storageRoot common.Hash) bool {
	it := db.NewIterator(rootSnapshotBlobKeyPrefix(storageRoot), nil)
	defer it.Release()
	return it.Next()
}

func WriteSnapshotMetadata(db ethdb.KeyValueWriter, metadata SnapshotMetadata) {
	enc, err := rlp.EncodeToBytes(metadata)
	if err != nil {
		log.Crit("Failed to encode snapshot metadata", "err", err)
	}
	if err := db.Put(snapshotMetadataKey(metadata.Address, metadata.BlockHash), enc); err != nil {
		log.Crit("Failed to write snapshot metadata", "err", err)
	}
}

func DeleteSnapshotMetadata(db ethdb.KeyValueWriter, address common.Address, blockHash common.Hash) {
	if err := db.Delete(snapshotMetadataKey(address, blockHash)); err != nil {
		log.Crit("Failed to delete snapshot metadata", "err", err)
	}
}

func ReadSnapshotMetadata(db ethdb.KeyValueReader, address common.Address, blockHash common.Hash) SnapshotMetadata {
	enc, err := db.Get(snapshotMetadataKey(address, blockHash))
	if err != nil {
		return SnapshotMetadata{}
	}
	var metadata SnapshotMetadata
	if err := rlp.DecodeBytes(enc, &metadata); err != nil {
		log.Crit("Failed to decode snapshot metadata", "err", err)
	}
	return metadata
}

func HasSnapshotMetadata(db ethdb.KeyValueReader, address common.Address, storageRoot common.Hash) bool {
	ok, _ := db.Has(snapshotMetadataKey(address, storageRoot))
	return ok
}

func IterateSnapshotBlobs(db ethdb.Iteratee) *SnapshotBlobIterator {
	return &SnapshotBlobIterator{
		Iterator: db.NewIterator(SnapshotBlobPrefix, nil),
	}
}

type SnapshotBlobIterator struct {
	ethdb.Iterator
}

func (it *SnapshotBlobIterator) StorageRoot() common.Hash {
	value := it.Key()
	var hash common.Hash
	copy(hash[:], value[len(SnapshotBlobPrefix):])
	return hash
}

func (it *SnapshotBlobIterator) Address() common.Address {
	value := it.Key()
	var address common.Address
	copy(address[:], value[len(SnapshotBlobPrefix)+common.HashLength:])
	return address
}

func (it *SnapshotBlobIterator) Blob() []byte {
	return it.Value()
}

func IterateSnapshotMetadata(db ethdb.Iteratee) *SnapshotMetadataIterator {
	return &SnapshotMetadataIterator{
		Iterator: db.NewIterator(SnapshotMetadataPrefix, nil),
	}
}

func IterateAccountSnapshotMetadata(db ethdb.Iteratee, address common.Address) *SnapshotMetadataIterator {
	return &SnapshotMetadataIterator{
		Iterator: db.NewIterator(accountSnapshotMetadataKeyPrefix(address), nil),
	}
}

type SnapshotMetadataIterator struct {
	ethdb.Iterator
}

func (it *SnapshotMetadataIterator) Address() common.Address {
	value := it.Key()
	var address common.Address
	copy(address[:], value[len(SnapshotMetadataPrefix):])
	return address
}

func (it *SnapshotMetadataIterator) BlockHash() common.Hash {
	value := it.Key()
	var hash common.Hash
	copy(hash[:], value[len(SnapshotMetadataPrefix)+common.AddressLength:])
	return hash
}

func (it *SnapshotMetadataIterator) Metadata() SnapshotMetadata {
	value := it.Value()
	var metadata SnapshotMetadata
	if err := rlp.DecodeBytes(value, &metadata); err != nil {
		log.Crit("Failed to decode snapshot metadata", "err", err)
	}
	return metadata
}
