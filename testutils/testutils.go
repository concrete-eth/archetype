package testutils

import (
	"math"
	"testing"

	"github.com/concrete-eth/archetype/testutils/gogen/archmod"
	"github.com/concrete-eth/archetype/testutils/gogen/datamod"
	archtypes "github.com/concrete-eth/archetype/types"
)

type ActionData_Add = archmod.ActionData_Add

func safeAddInt16(a, b int16) (int16, bool) {
	if b > 0 && a > math.MaxInt16-b {
		return 0, false
	}
	if b < 0 && a < math.MinInt16-b {
		return 0, false
	}
	return a + b, true
}

type TestCore struct {
	archtypes.BaseCore
}

var _ archmod.IActions = (*TestCore)(nil)

func (c *TestCore) setCounter(val int16) {
	counter := datamod.NewCounter(c.Datastore())
	counter.Get().Set(val)
}

func (c *TestCore) getCounter() int16 {
	counter := datamod.NewCounter(c.Datastore())
	return counter.Get().GetValue()
}

func (c *TestCore) add(summand int16) error {
	counter := c.getCounter()
	if res, ok := safeAddInt16(counter, summand); ok {
		c.setCounter(res)
	}
	return nil
}

func (c *TestCore) Add(action *archmod.ActionData_Add) error {
	return c.add(action.Summand)
}

func (c *TestCore) TicksPerBlock() uint {
	return 2
}

func (c *TestCore) Tick() {
	c.add(1)
}

func NewTestArchSpecs(t *testing.T) archtypes.ArchSpecs {
	return archtypes.ArchSpecs{
		Actions: archmod.ActionSpecs,
		Tables:  archmod.TableSpecs,
	}
}

func NewTestCore(t *testing.T) *TestCore {
	return &TestCore{}
}
