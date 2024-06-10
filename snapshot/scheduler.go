package snapshot

import (
	"encoding/binary"
	"sync"

	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

type Scheduler struct {
	db        ethdb.Database
	nonce     uint64
	schedules map[uint64]snapshot_types.Schedule
	lastRun   map[uint64]uint64
	lock      sync.RWMutex
}

func NewScheduler(db ethdb.Database, blockNumber uint64) *Scheduler {
	s := &Scheduler{
		db:        db,
		nonce:     0,
		schedules: make(map[uint64]snapshot_types.Schedule),
		lastRun:   make(map[uint64]uint64),
	}
	it := IterateSchedules(db)
	defer it.Release()
	for it.Next() {
		id := it.ID()
		schedule := it.Schedule()
		s.schedules[id] = schedule
		s.lastRun[id] = blockNumber
		if id >= s.nonce {
			s.nonce = id + 1
		}
	}
	return s
}

func (s *Scheduler) AddSchedule(schedule snapshot_types.Schedule, blockNumber uint64) (snapshot_types.ScheduleResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := s.nonce
	s.nonce++
	s.schedules[id] = schedule
	s.lastRun[id] = blockNumber
	WriteSchedule(s.db, id, schedule)
	log.Info("Added schedule", "id", id)
	return snapshot_types.ScheduleResponse{
		Schedule: schedule,
		ID:       id,
	}, nil
}

func (s *Scheduler) DeleteSchedule(id uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.schedules, id)
	delete(s.lastRun, id)
	DeleteSchedule(s.db, id)
	log.Info("Deleted schedule", "id", id)
	return nil
}

func (s *Scheduler) GetSchedules() (map[uint64]snapshot_types.Schedule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.schedules, nil
}

func (s *Scheduler) RunSchedule(blockNumber uint64, f func(schedule snapshot_types.Schedule)) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	run := false
	for id, schedule := range s.schedules {
		lastRun := s.lastRun[id]
		if (blockNumber - lastRun) >= schedule.BlockPeriod {
			f(schedule)
			s.lastRun[id] = blockNumber
			run = true
			log.Info("Ran schedule", "id", id)
		}
	}
	return run
}

func scheduleKey(id uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	return append(SchedulePrefix, buf...)
}

func WriteSchedule(db ethdb.KeyValueWriter, id uint64, schedule snapshot_types.Schedule) {
	enc, err := rlp.EncodeToBytes(schedule)
	if err != nil {
		log.Crit("Failed to encode schedule", "err", err)
	}
	if err := db.Put(scheduleKey(id), enc); err != nil {
		log.Crit("Failed to write schedule", "err", err)
	}
}

func ReadSchedule(db ethdb.KeyValueReader, id uint64) snapshot_types.Schedule {
	var schedule snapshot_types.Schedule
	enc, err := db.Get(scheduleKey(id))
	if err != nil {
		return schedule
	}
	err = rlp.DecodeBytes(enc, &schedule)
	if err != nil {
		log.Crit("Failed to decode schedule", "err", err)
	}
	return schedule
}

func DeleteSchedule(db ethdb.KeyValueWriter, id uint64) {
	if err := db.Delete(scheduleKey(id)); err != nil {
		log.Crit("Failed to delete schedule", "err", err)
	}
}

func IterateSchedules(db ethdb.Iteratee) *ScheduleIterator {
	return &ScheduleIterator{db.NewIterator(SchedulePrefix, nil)}
}

type ScheduleIterator struct {
	ethdb.Iterator
}

func (it *ScheduleIterator) ID() uint64 {
	return binary.BigEndian.Uint64(it.Key()[len(SchedulePrefix):])
}

func (it *ScheduleIterator) Schedule() snapshot_types.Schedule {
	var schedule snapshot_types.Schedule
	err := rlp.DecodeBytes(it.Value(), &schedule)
	if err != nil {
		log.Crit("Failed to decode schedule", "err", err)
	}
	return schedule
}
