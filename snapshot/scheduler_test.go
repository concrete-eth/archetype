package snapshot

import (
	"reflect"
	"testing"

	snapshot_types "github.com/concrete-eth/archetype/snapshot/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
)

var (
	testSchedule = snapshot_types.Schedule{
		Addresses: []common.Address{
			common.HexToAddress("0x1"),
			common.HexToAddress("0x2"),
			common.HexToAddress("0x3"),
		},
		BlockPeriod: 32,
		Replace:     true,
	}
)

func TestScheduleDB(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	id := uint64(1)
	if s := ReadSchedule(db, id); !reflect.DeepEqual(s, snapshot_types.Schedule{}) {
		t.Errorf("expected empty schedule, got %v", s)
	}
	WriteSchedule(db, id, testSchedule)
	if s := ReadSchedule(db, id); !reflect.DeepEqual(s, testSchedule) {
		t.Errorf("expected schedule %v, got %v", testSchedule, s)
	}
	DeleteSchedule(db, id)
	if s := ReadSchedule(db, id); !reflect.DeepEqual(s, snapshot_types.Schedule{}) {
		t.Errorf("expected empty schedule, got %v", s)
	}
}

func TestSchedulerAddReadDelete(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	blockNumber := uint64(0)
	scheduler := NewScheduler(db, blockNumber)
	r, err := scheduler.AddSchedule(testSchedule, blockNumber)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.ID != 0 {
		t.Errorf("expected ID 0, got %v", r.ID)
	}
	if !reflect.DeepEqual(r.Schedule, testSchedule) {
		t.Errorf("expected schedule %v, got %v", testSchedule, r.Schedule)
	}
	schedules, err := scheduler.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %v", len(schedules))
	}
	if !reflect.DeepEqual(schedules[r.ID], testSchedule) {
		t.Errorf("expected schedule %v, got %v", testSchedule, schedules[r.ID])
	}
	if err := scheduler.DeleteSchedule(r.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	schedules, err = scheduler.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules, got %v", len(schedules))
	}
}

func TestSchedulerCreation(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	blockNumber := uint64(0)
	id := uint64(0)
	WriteSchedule(db, id, testSchedule)
	scheduler := NewScheduler(db, blockNumber)
	schedules, err := scheduler.GetSchedules()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %v", len(schedules))
	}
	if !reflect.DeepEqual(schedules[0], testSchedule) {
		t.Errorf("expected schedule %v, got %v", testSchedule, schedules[0])
	}
}

func TestSchedulerRun(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	blockNumber := uint64(0)
	scheduler := NewScheduler(db, blockNumber)
	_, err := scheduler.AddSchedule(testSchedule, blockNumber)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	run := scheduler.RunSchedule(blockNumber, func(schedule snapshot_types.Schedule) {
		t.Error("expected no schedule run")
	})
	if run {
		t.Error("expected no schedule run")
	}
	called := false
	run = scheduler.RunSchedule(testSchedule.BlockPeriod, func(schedule snapshot_types.Schedule) {
		if !reflect.DeepEqual(schedule, schedule) {
			t.Errorf("expected schedule %v, got %v", schedule, schedule)
		}
		called = true
	})
	if !run {
		t.Error("expected schedule run")
	}
	if !called {
		t.Error("expected schedule run")
	}
}
