package snapshot

import (
	"bytes"
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/simulated"
	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/concrete-eth/archetype/snapshot/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

func NewTestSnapshotMaker() (*SnapshotMaker, SnapshotWriter, SnapshotReader, *simulated.SimulatedBackend) {
	gasLimit := uint64(1e7)
	sim := simulated.NewSimulatedBackend(core.GenesisAlloc{
		testSenderAddress: {Balance: math.MaxBig256},
	}, gasLimit, nil)
	m := NewSnapshotMaker(true)
	w := m.NewWriter(sim)
	r := m.NewReader(sim)
	return m, w, r, sim
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var (
	storageTesterCode = common.FromHex("608060405234801561001057600080fd5b5060b18061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c80633fa4f24514603757806355241077146051575b600080fd5b603f60005481565b60405190815260200160405180910390f35b6061605c3660046063565b600055565b005b600060208284031215607457600080fd5b503591905056fea2646970667358221220a1490c138af3ca93bfdcc87a46306f71fe97efe367215b8733eedd154c0dd4e164736f6c634300080f0033")
	valueSig          = crypto.Keccak256([]byte("value()"))[:4]
	setSig            = crypto.Keccak256([]byte("setValue(uint256)"))[:4]
)

func deployStorageTester(t *testing.T, sim *simulated.SimulatedBackend) common.Address {
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	checkError(t, err)
	tx := types.NewContractCreation(0, common.Big0, 1e6, gasPrice, storageTesterCode)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, testKey)
	checkError(t, err)
	err = sim.SendTransaction(context.Background(), tx)
	checkError(t, err)
	sim.Commit()
	receipt, err := sim.TransactionReceipt(context.Background(), tx.Hash())
	checkError(t, err)
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("expected successful receipt, got %v", receipt.Status)
	}
	return receipt.ContractAddress
}

func setValue(t *testing.T, sim *simulated.SimulatedBackend, address common.Address, value *big.Int) {
	gasPrice, err := sim.SuggestGasPrice(context.Background())
	checkError(t, err)
	nonce, err := sim.NonceAt(context.Background(), testSenderAddress, nil)
	checkError(t, err)
	tx := types.NewTransaction(nonce, address, common.Big0, 1e6, gasPrice, append(
		setSig,
		common.BigToHash(value).Bytes()...,
	))
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, testKey)
	checkError(t, err)
	err = sim.SendTransaction(context.Background(), tx)
	checkError(t, err)
	sim.Commit()
}

func getValue(t *testing.T, sim *simulated.SimulatedBackend, address common.Address) *big.Int {
	ret, err := sim.CallContract(context.Background(), ethereum.CallMsg{
		From: testSenderAddress,
		To:   &address,
		Data: valueSig,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := new(big.Int).SetBytes(ret)
	return n
}

func checkMetadata(t *testing.T, m SnapshotMetadataWithStatus, address common.Address, blockHash common.Hash, blockNumber uint64, status SnapshotStatus) {
	if m.Address != address {
		t.Errorf("expected address %v, got %v", address, m.Address)
	}
	if m.BlockHash != blockHash {
		t.Errorf("expected block hash %v, got %v", blockHash, m.BlockHash)
	}
	if m.BlockNumber.Cmp(new(big.Int).SetUint64(blockNumber)) != 0 {
		t.Errorf("expected block number %v, got %v", blockNumber, m.BlockNumber)
	}
	if m.Status != status {
		t.Errorf("expected status %v, got %v", status, m.Status)
	}
}

func waitForSnapshot(t *testing.T, reader SnapshotReader, address common.Address, blockHash common.Hash) SnapshotResponse {
	startTime := time.Now()
	for {
		res, err := reader.Get(address, blockHash)
		checkError(t, err)
		if res.Status == SnapshotStatus_Done {
			return res
		} else if res.Status == SnapshotStatus_Pending {
			if time.Since(startTime) > 1*time.Second {
				t.Fatal("timeout")
				break
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			t.Fatalf("unexpected status %v", res.Status)
		}
	}
	return SnapshotResponse{}
}

func TestSnapshot(t *testing.T) {
	// log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	_, writer, reader, sim := NewTestSnapshotMaker()
	rw := writer.(*snapshotReaderWriter)
	sim.Commit()

	address := common.HexToAddress("0x1")
	blockHash := sim.BlockChain().CurrentBlock().Hash()

	// Get non-existent block
	_, err := reader.Get(address, common.Hash{1})
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
	// Get non-existent snapshot for existing block
	_, err = reader.Get(address, blockHash)
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
	// Get the last out of zero snapshots
	_, err = reader.Last(address)
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
	// Get all out of zero snapshots
	l, err := reader.List(address)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(l) != 0 {
		t.Errorf("expected no snapshots, got %v", l)
	}

	// New snapshot
	mm, err := writer.New([]common.Address{address}, blockHash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mm) != 1 {
		t.Errorf("expected 1 metadata, got %v", len(mm))
	}
	checkMetadata(t, mm[0], address, blockHash, 1, SnapshotStatus_Pending)
	_, err = reader.Last(address)
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
	rw.runSnapshotWorkerTask(<-rw.taskQueueChan)
	m, err := reader.Last(address)
	checkError(t, err)
	checkMetadata(t, m, address, blockHash, 1, SnapshotStatus_Done)

	// New block
	// Note there are no storage changes in this block
	previousBlockHash := blockHash
	sim.Commit()
	blockHash = sim.BlockChain().CurrentBlock().Hash()

	// Update snapshot
	mm, err = writer.Update([]common.Address{address}, blockHash)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(mm) != 1 {
		t.Errorf("expected 1 metadata, got %v", len(mm))
	}
	checkMetadata(t, mm[0], address, blockHash, 2, SnapshotStatus_Done)

	// Get the previous snapshot
	res, err := reader.Get(address, previousBlockHash)
	if err != ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}

	// Deploy storage tester contract
	// A new block is created
	contractAddress := deployStorageTester(t, sim)

	// Set value
	// A new block is created
	value := common.Big2
	setValue(t, sim, contractAddress, value)
	if v := getValue(t, sim, contractAddress); v.Cmp(value) != 0 {
		t.Errorf("expected value %v, got %v", value, v)
	}

	blockHash = sim.BlockChain().CurrentBlock().Hash()

	// New snapshot
	mm, err = writer.New([]common.Address{contractAddress}, blockHash)
	checkError(t, err)
	if len(mm) != 1 {
		t.Errorf("expected 1 metadata, got %v", len(mm))
	}
	checkMetadata(t, mm[0], contractAddress, blockHash, 4, SnapshotStatus_Pending)

	// Wait for snapshot
	rw.runSnapshotWorkerTask(<-rw.taskQueueChan)
	res, err = reader.Get(contractAddress, blockHash)
	checkError(t, err)
	checkMetadata(t, res.SnapshotMetadataWithStatus, contractAddress, blockHash, 4, SnapshotStatus_Done)

	// Check blob
	rawBlob, err := utils.Decompress(res.Storage)
	checkError(t, err)
	it := utils.BlobToStorageIt(rawBlob)
	storage := make(map[common.Hash][]byte)
	for it.Next() {
		slot := it.Hash()
		value := it.Slot()
		storage[slot] = value
	}
	if len(storage) != 1 {
		t.Errorf("expected 1 storage slot, got %v", len(storage))
	}
	for _, v := range storage {
		dec, err := utils.DecodeSnapshotSlot(v)
		checkError(t, err)
		if dec != common.BigToHash(value) {
			t.Errorf("expected value %v, got %v", value, dec)
		}
		break
	}
}

func TestSchedulerReadWrite(t *testing.T) {
	_, writer, reader, _ := NewTestSnapshotMaker()

	// Get schedules when there are none
	s, err := reader.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(s) != 0 {
		t.Errorf("expected no schedules, got %v", len(s))
	}

	// Add schedule
	schedule := snapshot_types.Schedule{BlockPeriod: 1}
	res, err := writer.AddSchedule(schedule)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ID != 0 {
		t.Errorf("expected ID 0, got %v", res.ID)
	}
	if !reflect.DeepEqual(res.Schedule, schedule) {
		t.Errorf("expected schedule %v, got %v", schedule, res.Schedule)
	}
	s, err = reader.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(s) != 1 {
		t.Errorf("expected 1 schedule, got %v", len(s))
	}
	if !reflect.DeepEqual(s[0], res.Schedule) {
		t.Errorf("expected schedule %v, got %v", res, s[0])
	}

	// Delete schedule
	err = writer.DeleteSchedule(res.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	s, err = reader.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(s) != 0 {
		t.Errorf("expected no schedules, got %v", len(s))
	}

	// Add schedule
	res, err = writer.AddSchedule(schedule)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ID != 1 {
		t.Errorf("expected ID 1, got %v", res.ID)
	}

	writer.AddSchedule(schedule)
	writer.AddSchedule(schedule)
	writer.AddSchedule(schedule)

	res, err = writer.AddSchedule(schedule)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ID != 5 {
		t.Errorf("expected ID 5, got %v", res.ID)
	}
}

func TestScheduler(t *testing.T) {
	// log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	maker, writer, reader, sim := NewTestSnapshotMaker()
	go maker.RunSnapshotWorker()
	// go maker.RunScheduler()
	sim.Commit()

	contractAddress := deployStorageTester(t, sim)
	setValue(t, sim, contractAddress, common.Big2)

	var blockHash common.Hash
	var err error

	// New snapshots
	blockHash = sim.BlockChain().CurrentBlock().Hash()
	_, err = writer.New([]common.Address{contractAddress}, blockHash)
	checkError(t, err)
	sim.Commit()

	// Wait for snapshots
	l, err := reader.List(contractAddress)
	checkError(t, err)
	if len(l) != 1 {
		t.Errorf("expected 1 snapshots, got %v", len(l))
	}

	// Add schedule
	schedule := snapshot_types.Schedule{
		Addresses:   []common.Address{contractAddress},
		BlockPeriod: 2,
		Replace:     false,
	}
	res, err := writer.AddSchedule(schedule)
	checkError(t, err)
	id := res.ID

	// New block
	sim.Commit()
	// Trigger scheduler
	// Schedule should not run as block period has not elapsed
	writer.(*snapshotReaderWriter).runScheduler()

	l, err = reader.List(contractAddress)
	checkError(t, err)
	if len(l) != 1 {
		t.Errorf("expected 1 snapshots, got %v", len(l))
	}

	// New block
	sim.Commit()
	// Trigger scheduler
	// Schedule should run as block period has elapsed
	writer.(*snapshotReaderWriter).runScheduler()

	l, err = reader.List(contractAddress)
	checkError(t, err)
	if len(l) != 2 {
		t.Errorf("expected 2 snapshots, got %v", len(l))
	}

	// Delete schedule
	err = writer.DeleteSchedule(id)
	checkError(t, err)

	// Add schedule with replacement
	schedule.Replace = true
	res, err = writer.AddSchedule(schedule)
	checkError(t, err)
	// id = res.ID

	// Two blocks
	// sim.Commit()
	setValue(t, sim, contractAddress, common.Big3)
	sim.Commit()
	// Trigger scheduler
	writer.(*snapshotReaderWriter).runScheduler()

	blockHash = sim.BlockChain().CurrentBlock().Hash()
	waitForSnapshot(t, reader, contractAddress, blockHash)

	l, err = reader.List(contractAddress)
	checkError(t, err)
	if len(l) != 2 {
		t.Errorf("expected 2 snapshots, got %v", len(l))
	}

	currentBlockNumber := sim.BlockChain().CurrentBlock().Number.Uint64()
	previousSnapshotBlockNumber := currentBlockNumber - 2

	// Check list contains latest two snapshots
	if l[0].BlockNumber.Cmp(l[1].BlockNumber) == 0 {
		t.Errorf("expected different block numbers, got %v", l)
	}
	if l0 := l[0].BlockNumber.Uint64(); l0 != currentBlockNumber && l0 != previousSnapshotBlockNumber {
		t.Errorf("expected block number %v or %v, got %v", currentBlockNumber, previousSnapshotBlockNumber, l0)
	}
	if l1 := l[1].BlockNumber.Uint64(); l1 != currentBlockNumber && l1 != previousSnapshotBlockNumber {
		t.Errorf("expected block number %v or %v, got %v", currentBlockNumber, previousSnapshotBlockNumber, l1)
	}
}