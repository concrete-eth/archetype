package snapshot

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/simulated"
	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/concrete-eth/archetype/snapshot/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompression(t *testing.T) {
	data := []byte("test data")
	compressed, err := utils.Compress(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	decompressed, err := utils.Decompress(compressed)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !bytes.Equal(decompressed, data) {
		t.Errorf("expected %v, got %v", data, decompressed)
	}
}

func TestMetadataDB(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	metadata := SnapshotMetadata{
		Address:     common.HexToAddress("0x1"),
		BlockHash:   common.HexToHash("0x2"),
		BlockNumber: common.Big3,
		StorageRoot: common.HexToHash("0x4"),
	}
	if ok := HasSnapshotMetadata(db, metadata.Address, metadata.BlockHash); ok {
		t.Errorf("expected no metadata, got %v", ok)
	}
	WriteSnapshotMetadata(db, metadata)
	if ok := HasSnapshotMetadata(db, metadata.Address, metadata.BlockHash); !ok {
		t.Errorf("expected metadata")
	}
	if m := ReadSnapshotMetadata(db, metadata.Address, metadata.BlockHash); !reflect.DeepEqual(m, metadata) {
		t.Errorf("expected metadata %v, got %v", metadata, m)
	}
	DeleteSnapshotMetadata(db, metadata.Address, metadata.BlockHash)
	if ok := HasSnapshotMetadata(db, metadata.Address, metadata.BlockHash); ok {
		t.Errorf("expected no metadata, got %v", ok)
	}
}

func TestBlobDB(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	address := common.HexToAddress("0x1")
	blockHash := common.HexToHash("0x2")
	blob := []byte("test data")
	if ok := HasSnapshotBlob(db, address, blockHash); ok {
		t.Errorf("expected no blob, got %v", ok)
	}
	WriteSnapshotBlob(db, address, blockHash, blob)
	if ok := HasSnapshotBlob(db, address, blockHash); !ok {
		t.Errorf("expected blob")
	}
	if b := ReadSnapshotBlob(db, address, blockHash); !bytes.Equal(b, blob) {
		t.Errorf("expected blob %v, got %v", blob, b)
	}
	DeleteSnapshotBlob(db, address, blockHash)
	if ok := HasSnapshotBlob(db, address, blockHash); ok {
		t.Errorf("expected no blob, got %v", ok)
	}
}

type testStorageIterator struct {
	data  [][2]common.Hash
	index int
}

var _ snapshot.StorageIterator = (*testStorageIterator)(nil)

func newTestStorageIterator(data [][2]common.Hash) *testStorageIterator {
	return &testStorageIterator{
		data:  data,
		index: -1,
	}
}

func (it *testStorageIterator) Next() bool {
	it.index++
	return it.index < len(it.data)
}

func (it *testStorageIterator) Hash() common.Hash {
	return it.data[it.index][0]
}

func (it *testStorageIterator) Slot() []byte {
	enc, _ := utils.EncodeSnapshotSlot(it.data[it.index][1])
	return enc
}

func (it *testStorageIterator) Error() error {
	return nil
}

func (it *testStorageIterator) Release() {}

func TestStorageBlob(t *testing.T) {
	data := [][2]common.Hash{
		{common.HexToHash("0x1"), common.HexToHash("0x2")},
		{common.HexToHash("0x3"), common.HexToHash("0x4")},
	}
	storageIt := newTestStorageIterator(data)
	blob, err := utils.StorageItToBlob(storageIt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	readStorageIt := utils.BlobToStorageIt(blob)
	readStorage := make(map[common.Hash][]byte)
	for readStorageIt.Next() {
		slot := readStorageIt.Hash()
		value := readStorageIt.Slot()
		readStorage[slot] = value
	}
	for _, pair := range data {
		slot := pair[0]
		value, _ := utils.EncodeSnapshotSlot(pair[1])
		if !bytes.Equal(readStorage[slot], value) {
			t.Errorf("expected %v, got %v", value, readStorage[slot])
		}
	}
}

var (
	testKey, _        = crypto.HexToECDSA("d17bd946feb884d463d58fb702b94dd0457ca349338da1d732a57856cf777ccd") // 0xCcca11AbAC28D9b6FceD3a9CA73C434f6b33B215
	testSenderAddress = crypto.PubkeyToAddress(testKey.PublicKey)
)

func NewTestSnapshotMakerWithConcrete(registry concrete.PrecompileRegistry) (*SnapshotMaker, SnapshotWriter, SnapshotReader, *simulated.SimulatedBackend) {
	gasLimit := uint64(1e7)
	sim := simulated.NewSimulatedBackend(types.GenesisAlloc{
		testSenderAddress: {Balance: math.MaxBig256},
	}, gasLimit, registry)
	m := NewSnapshotMaker(true)
	w := m.NewWriter(sim)
	r := m.NewReader(sim)
	return m, w, r, sim
}

func NewTestSnapshotMaker() (*SnapshotMaker, SnapshotWriter, SnapshotReader, *simulated.SimulatedBackend) {
	return NewTestSnapshotMakerWithConcrete(nil)
}

type storageSetterPc struct {
	lib.BlankPrecompile
}

func (p *storageSetterPc) Run(API api.Environment, input []byte) ([]byte, error) {
	key := common.BytesToHash(input[:32])
	value := common.BytesToHash(input[32:])
	API.StorageStore(key, value)
	return nil, nil
}

type testRegistry struct {
	addresses []common.Address
}

var _ concrete.PrecompileRegistry = (*testRegistry)(nil)

func (r *testRegistry) Precompile(addr common.Address, _ uint64) (concrete.Precompile, bool) {
	for _, a := range r.addresses {
		if a == addr {
			return &storageSetterPc{}, true
		}
	}
	return nil, false
}

func (r *testRegistry) Precompiles(_ uint64) concrete.PrecompileMap {
	m := make(concrete.PrecompileMap)
	for _, addr := range r.addresses {
		m[addr] = &storageSetterPc{}
	}
	return m
}

func (r *testRegistry) PrecompiledAddresses(_ uint64) []common.Address {
	return r.addresses
}

func (r *testRegistry) PrecompiledAddressesSet(_ uint64) map[common.Address]struct{} {
	set := make(map[common.Address]struct{})
	for _, addr := range r.addresses {
		set[addr] = struct{}{}
	}
	return set
}

func sendSetValueTx(t *testing.T, sim *simulated.SimulatedBackend, addr common.Address, key, value common.Hash) {
	r := require.New(t)

	nonce, err := sim.PendingNonceAt(context.Background(), testSenderAddress)
	r.NoError(err)
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	r.NoError(err)
	gas := uint64(1e6)

	input := make([]byte, 64)
	copy(input[32-len(key.Bytes()):], key.Bytes())
	copy(input[64-len(value.Bytes()):], value.Bytes())

	tx := types.NewTransaction(nonce, addr, common.Big0, gas, gasPrice, input)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)
	r.NoError(err)

	err = sim.SendTransaction(context.Background(), signedTx)
	r.NoError(err)

	_, pending, err := sim.TransactionByHash(context.Background(), signedTx.Hash())
	r.NoError(err)
	r.True(pending)
}

func TestSnapshot(t *testing.T) {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, true)))

	var (
		r         = require.New(t)
		addr1     = common.HexToAddress("0x12340001")
		addr2     = common.HexToAddress("0x12340002")
		addr3     = common.HexToAddress("0x12340003")
		addresses = []common.Address{addr1, addr2, addr3}
	)

	registry := &testRegistry{addresses: addresses}
	_, writer, reader, sim := NewTestSnapshotMakerWithConcrete(registry)
	rw := writer.(*snapshotReaderWriter)

	sim.Commit()
	block := sim.BlockChain().CurrentBlock()

	// Query non-existent snapshots
	for _, addr := range addresses {
		// Get non-existent block
		_, err := reader.Get(addr, common.Hash{1})
		r.Equal(ErrSnapshotNotFound, err)
		// Get non-existent snapshot for existing block
		_, err = reader.Get(addr, block.Hash())
		r.Equal(ErrSnapshotNotFound, err)
		// Get the last out of zero snapshots
		_, err = reader.Last(addr)
		r.Equal(ErrSnapshotNotFound, err)
		// Get all out of zero snapshots
		l, err := reader.List(addr)
		r.NoError(err)
		r.Len(l, 0)
	}

	// Create snapshots
	mm, err := writer.New(SnapshotQuery{Addresses: addresses, BlockHash: block.Hash()})
	r.NoError(err)
	r.Len(mm, len(addresses))
	checked := make(map[common.Address]struct{})
	for _, m := range mm {
		r.NotContains(checked, m.Address)
		r.Contains(addresses, m.Address)
		checked[m.Address] = struct{}{}

		// Check metadata
		r.Equal(block.Hash(), m.BlockHash)
		r.Equal(block.Number, m.BlockNumber)
		r.Equal(SnapshotStatus_Pending, m.Status)

		// Get the snapshot
		got, err := reader.Get(m.Address, block.Hash())
		r.NoError(err)
		r.Equal(m, got.SnapshotMetadataWithStatus)

		// Get the last snapshot (not found since still pending)
		_, err = reader.Last(addr1)
		r.Equal(ErrSnapshotNotFound, err)

		// Get all snapshots
		list, err := reader.List(m.Address)
		r.NoError(err)
		r.Len(list, 1)
		r.Equal(m, list[0])
	}
	r.Len(checked, len(addresses))

	// Run snapshot worker
	timeout := time.After(1 * time.Second)
	for i := 0; i < len(addresses); i++ {
		select {
		case <-timeout:
			t.Fatal("timeout")
		case task := <-rw.taskQueueChan:
			r.Contains(addresses, task.Metadata.Address)
			rw.runSnapshotWorkerTask(task)
		}
	}
	select {
	case <-rw.taskQueueChan:
		t.Fatal("expected no more tasks")
	default:
	}

	// Check snapshots
	for _, addr := range addresses {
		m, err := reader.Last(addr)
		r.NoError(err)
		r.Equal(block.Hash(), m.BlockHash)
		r.Equal(block.Number, m.BlockNumber)
		r.Equal(SnapshotStatus_Done, m.Status)
	}

	// New block
	// Note there are no storage changes in this block
	sim.Commit()
	block = sim.BlockChain().CurrentBlock()
	prevBlock := sim.BlockChain().GetHeaderByHash(block.ParentHash)

	// Update snapshots
	mm, err = writer.Update(SnapshotQuery{Addresses: addresses, BlockHash: block.Hash()})
	r.NoError(err)
	r.Len(mm, len(addresses))
	checked = make(map[common.Address]struct{})
	for _, m := range mm {
		r.NotContains(checked, m.Address)
		r.Contains(addresses, m.Address)
		checked[m.Address] = struct{}{}

		// Check metadata
		r.Equal(block.Hash(), m.BlockHash)
		r.Equal(block.Number, m.BlockNumber)
		r.Equal(SnapshotStatus_Done, m.Status)

		// Get the previous snapshot
		_, err = reader.Get(m.Address, prevBlock.Hash())
		r.Equal(ErrSnapshotNotFound, err)

		// Get the snapshot
		got, err := reader.Get(m.Address, block.Hash())
		r.NoError(err)
		r.Equal(m, got.SnapshotMetadataWithStatus)

		// Get the last snapshot
		last, err := reader.Last(m.Address)
		r.NoError(err)
		r.Equal(m, last)

		// Get all snapshots
		list, err := reader.List(m.Address)
		r.NoError(err)
		r.Len(list, 1)
		r.Equal(m, list[0])
	}
	r.Len(checked, len(addresses))

	// Set storage
	for ii, addr := range addresses {
		value := big.NewInt(int64(ii + 1))
		sendSetValueTx(t, sim, addr, common.Hash{}, common.BigToHash(value))
	}

	// New block
	sim.Commit()
	block = sim.BlockChain().CurrentBlock()
	prevBlock = sim.BlockChain().GetHeaderByHash(block.ParentHash)

	// New snapshots with different storage
	mm, err = writer.New(SnapshotQuery{Addresses: addresses, BlockHash: block.Hash()})
	r.NoError(err)
	r.Len(mm, len(addresses))
	checked = make(map[common.Address]struct{})
	for _, m := range mm {
		r.NotContains(checked, m.Address)
		r.Contains(addresses, m.Address)
		checked[m.Address] = struct{}{}

		// Check metadata
		r.Equal(block.Hash(), m.BlockHash)
		r.Equal(block.Number, m.BlockNumber)
		r.Equal(SnapshotStatus_Pending, m.Status)

		// Get the previous snapshot
		prev, err := reader.Get(m.Address, prevBlock.Hash())
		r.NoError(err)
		r.Equal(prevBlock.Hash(), prev.BlockHash)
		r.Equal(prevBlock.Number, prev.BlockNumber)
		r.Equal(SnapshotStatus_Done, prev.Status)
		r.NotEqual(m.StorageRoot, prev.StorageRoot)

		// Get the snapshot
		got, err := reader.Get(m.Address, block.Hash())
		r.NoError(err)
		r.Equal(m, got.SnapshotMetadataWithStatus)

		// Get the last snapshot (will be from previous block since new block is pending)
		last, err := reader.Last(m.Address)
		r.NoError(err)
		r.Equal(prev.SnapshotMetadataWithStatus, last)

		// Get all snapshots
		list, err := reader.List(m.Address)
		r.NoError(err)
		r.Len(list, 2)

		if !(assert.ObjectsAreEqual(m, list[0]) && assert.ObjectsAreEqual(prev.SnapshotMetadataWithStatus, list[1])) &&
			!(assert.ObjectsAreEqual(m, list[1]) && assert.ObjectsAreEqual(prev.SnapshotMetadataWithStatus, list[0])) {
			t.Fatalf("expected %v and %v, got %v", m, prev.SnapshotMetadataWithStatus, list)
		}
	}

	// Run snapshot worker
	timeout = time.After(1 * time.Second)
	for i := 0; i < len(addresses); i++ {
		select {
		case <-timeout:
			t.Fatal("timeout")
		case task := <-rw.taskQueueChan:
			r.Contains(addresses, task.Metadata.Address)
			rw.runSnapshotWorkerTask(task)
		}
	}
	select {
	case <-rw.taskQueueChan:
		t.Fatal("expected no more tasks")
	default:
	}

	// Check snapshots
	for ii, addr := range addresses {
		res, err := reader.Get(addr, block.Hash())
		r.NoError(err)
		r.Equal(SnapshotStatus_Done, res.Status)
		// Check storage
		rawBlob, err := utils.Decompress(res.Storage)
		r.NoError(err)
		it := utils.BlobToStorageIt(rawBlob)
		storage := make(map[common.Hash][]byte)
		for it.Next() {
			slot := it.Hash()
			value := it.Slot()
			storage[slot] = value
		}
		r.Len(storage, 1)
		for k, v := range storage {
			r.Equal(crypto.Keccak256Hash(common.Hash{}.Bytes()), k)
			dec, err := utils.DecodeSnapshotSlot(v)
			r.NoError(err)
			value := big.NewInt(int64(ii + 1))
			r.Equal(common.BigToHash(value), dec)
			break
		}
	}
}

func TestMakeBlobFromScratch(t *testing.T) {
	r := require.New(t)
	addr := common.HexToAddress("0x12340001")
	registry := &testRegistry{addresses: []common.Address{addr}}
	_, writer, _, sim := NewTestSnapshotMakerWithConcrete(registry)
	rw := writer.(*snapshotReaderWriter)

	storage := make(map[common.Hash]common.Hash)
	for i := 0; i < 10; i++ {
		key := common.BigToHash(big.NewInt(int64(i)))
		value := common.BigToHash(big.NewInt(int64(i + 1)))
		storage[crypto.Keccak256Hash(key.Bytes())] = value
		sendSetValueTx(t, sim, addr, key, value)
	}

	sim.Commit()
	block := sim.BlockChain().CurrentBlock()

	rawBlob, err := rw.makeBlobFromScratch(addr, block.Root)
	r.NoError(err)
	blob, err := utils.Decompress(rawBlob)
	r.NoError(err)

	it := utils.BlobToStorageIt(blob)
	readStorage := make(map[common.Hash]common.Hash)
	for it.Next() {
		slot := it.Hash()
		value := it.Slot()
		readStorage[slot] = common.BytesToHash(value)
	}
	r.Equal(storage, readStorage)
}

func TestMakeBlobFromPrevious(t *testing.T) {
	var (
		r                      = require.New(t)
		addr                   = common.HexToAddress("0x12340001")
		registry               = &testRegistry{addresses: []common.Address{addr}}
		_, writer, reader, sim = NewTestSnapshotMakerWithConcrete(registry)
		rw                     = writer.(*snapshotReaderWriter)
	)

	storage := make(map[common.Hash]common.Hash)
	setKv := func(key, value int) {
		keyHash := common.BigToHash(big.NewInt(int64(key)))
		valueHash := common.BigToHash(big.NewInt(int64(value)))
		sendSetValueTx(t, sim, addr, keyHash, valueHash)
		storage[crypto.Keccak256Hash(keyHash.Bytes())] = valueHash
	}

	for i := 0; i < 10; i++ {
		setKv(i, i+1)
	}

	sim.Commit()
	block := sim.BlockChain().CurrentBlock()

	_, err := writer.New(SnapshotQuery{Addresses: []common.Address{addr}, BlockHash: block.Hash()})
	r.NoError(err)

	rw.runSnapshotWorkerTask(<-rw.taskQueueChan)

	got, err := reader.Last(addr)
	r.NoError(err)
	r.Equal(block.Hash(), got.BlockHash)
	r.Equal(SnapshotStatus_Done, got.Status)

	for i := 10; i < 20; i++ {
		setKv(i, i+1)
	}

	for i := 0; i < 256; i++ {
		sim.Commit()
	}

	for i := 20; i < 30; i++ {
		setKv(i, i+1)
	}
	for i := 0; i < 5; i++ {
		setKv(i, i+10)
	}

	sim.Commit()
	block = sim.BlockChain().CurrentBlock()

	rawBlob, err := rw.makeBlob(addr, block.Hash(), block.Root)
	r.NoError(err)
	blob, err := utils.Decompress(rawBlob)
	r.NoError(err)

	it := utils.BlobToStorageIt(blob)
	readStorage := make(map[common.Hash]common.Hash)
	for it.Next() {
		slot := it.Hash()
		value := it.Slot()
		readStorage[slot] = common.BytesToHash(value)
	}
	r.Equal(storage, readStorage)
}

func TestSchedulerReadWrite(t *testing.T) {
	r := require.New(t)
	_, writer, reader, _ := NewTestSnapshotMaker()

	// Get schedules when there are none
	s, err := reader.GetSchedules()
	r.NoError(err)
	r.Len(s, 0)

	// Add schedule
	schedule := snapshot_types.Schedule{BlockPeriod: 1}
	res, err := writer.AddSchedule(schedule)
	r.NoError(err)
	r.Equal(uint64(0), res.ID)
	r.Equal(schedule, res.Schedule)

	// Get schedules
	s, err = reader.GetSchedules()
	r.NoError(err)
	r.Len(s, 1)
	r.Equal(res.Schedule, s[0])

	// Delete schedule
	err = writer.DeleteSchedule(res.ID)
	r.NoError(err)

	// Get schedules when there are none
	s, err = reader.GetSchedules()
	r.NoError(err)
	r.Len(s, 0)

	// Add schedules
	for i := 0; i < 3; i++ {
		res, err = writer.AddSchedule(schedule)
		r.NoError(err)
		r.Equal(uint64(i+1), res.ID)
		r.Equal(schedule, res.Schedule)
	}

	res, err = writer.AddSchedule(schedule)
	r.NoError(err)
	r.Equal(uint64(4), res.ID)
	r.Equal(schedule, res.Schedule)
}

func TestScheduler(t *testing.T) {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, true)))

	var (
		r         = require.New(t)
		addr1     = common.HexToAddress("0x12340001")
		addresses = []common.Address{addr1}
	)

	registry := &testRegistry{addresses: addresses}
	_, writer, reader, sim := NewTestSnapshotMakerWithConcrete(registry)
	rw := writer.(*snapshotReaderWriter)

	sim.Commit()
	block := sim.BlockChain().CurrentBlock()
	blockHash1 := block.Hash()

	// New snapshot
	_, err := writer.New(SnapshotQuery{Addresses: addresses, BlockHash: block.Hash()})
	r.NoError(err)

	// Wait for snapshots
	for _, addr := range addresses {
		list, err := reader.List(addr)
		r.NoError(err)
		r.Len(list, 1)
		r.Equal(blockHash1, list[0].BlockHash)
	}

	// Add schedule
	res, err := writer.AddSchedule(Schedule{
		Addresses:   addresses,
		BlockPeriod: 2,
		Replace:     false,
	})
	r.NoError(err)
	id := res.ID

	// New block
	sim.Commit()
	// Trigger scheduler
	// Schedule should not run as block period has not elapsed
	rw.runScheduler()

	for _, addr := range addresses {
		list, err := reader.List(addr)
		r.NoError(err)
		r.Len(list, 1)
	}

	// New block
	sim.Commit()
	block = sim.BlockChain().CurrentBlock()
	blockHash3 := block.Hash()
	// Trigger scheduler
	// Schedule should run as block period has elapsed
	rw.runScheduler()

	for _, addr := range addresses {
		list, err := reader.List(addr)
		r.NoError(err)
		r.Len(list, 2)
		r.ElementsMatch([]common.Hash{blockHash1, blockHash3}, []common.Hash{list[0].BlockHash, list[1].BlockHash})
	}

	// Delete schedule
	err = writer.DeleteSchedule(id)
	r.NoError(err)

	// Add schedule with replacement
	res, err = writer.AddSchedule(Schedule{
		Addresses:   addresses,
		BlockPeriod: 2,
		Replace:     true,
	})
	r.NoError(err)

	// Two new blocks
	sim.Commit()
	sim.Commit()
	block = sim.BlockChain().CurrentBlock()
	blockHash4 := block.Hash()
	// Trigger scheduler
	rw.runScheduler()

	// Run snapshot worker
	timeout := time.After(1 * time.Second)
	for i := 0; i < len(addresses)*3; i++ {
		select {
		case <-timeout:
			t.Fatal("timeout")
		case task := <-rw.taskQueueChan:
			r.Contains(addresses, task.Metadata.Address)
			rw.runSnapshotWorkerTask(task)
		}
	}
	select {
	case <-rw.taskQueueChan:
		t.Fatal("expected no more tasks")
	default:
	}

	for _, addr := range addresses {
		timeout := time.After(1 * time.Second)
		for {
			select {
			case <-timeout:
				t.Fatal("timeout")
			default:
			}
			got, err := reader.Last(addr)
			r.NoError(err)
			r.Equal(blockHash4, got.BlockHash)
			if got.Status == SnapshotStatus_Done {
				break
			}
		}
	}

	for _, addr := range addresses {
		list, err := reader.List(addr)
		r.NoError(err)
		r.Len(list, 2)
		r.ElementsMatch([]common.Hash{blockHash3, blockHash4}, []common.Hash{list[0].BlockHash, list[1].BlockHash})
	}
}
