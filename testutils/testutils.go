package testutils

import (
	"math"
	"testing"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/testutils/gogen/archmod"
	"github.com/concrete-eth/archetype/testutils/gogen/datamod"
)

type ActionData_Add = archmod.ActionData_Add
type RowData_Counter = archmod.RowData_Counter

func safeAddInt16(a, b int16) (int16, bool) {
	if b > 0 && a > math.MaxInt16-b {
		return 0, false
	}
	if b < 0 && a < math.MinInt16-b {
		return 0, false
	}
	return a + b, true
}

type Core struct {
	arch.BaseCore
}

var _ archmod.IActions = (*Core)(nil)

func (c *Core) SetCounter(val int16) {
	counter := datamod.NewCounter(c.Datastore())
	counter.Get().Set(val)
}

func (c *Core) GetCounter() int16 {
	counter := datamod.NewCounter(c.Datastore())
	return counter.Get().GetValue()
}

func (c *Core) add(summand int16) error {
	counter := c.GetCounter()
	if res, ok := safeAddInt16(counter, summand); ok {
		c.SetCounter(res)
	}
	return nil
}

func (c *Core) mul(factor int16) error {
	counter := c.GetCounter()
	c.SetCounter(counter * factor)
	return nil
}

func (c *Core) Add(action *archmod.ActionData_Add) error {
	return c.add(action.Summand)
}

func (c *Core) TicksPerBlock() uint {
	return 2
}

func (c *Core) Tick() {
	c.mul(2)
}

func NewTestArchSpecs(t *testing.T) arch.ArchSchemas {
	return arch.ArchSchemas{
		Actions: archmod.ActionSpecs,
		Tables:  archmod.TableSpecs,
	}
}

func NewTestCore(t *testing.T) *Core {
	return &Core{}
}
