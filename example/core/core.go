package core

import (
	"errors"
	"math"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/gogen/datamod"
	"github.com/concrete-eth/archetype/utils"
	"golang.org/x/exp/constraints"
)

const (
	G_NUMERATOR          int32 = 400
	G_DENOMINATOR        int32 = 1
	INTERVAL_NUMERATOR   int16 = 1
	INTERVAL_DENOMINATOR int16 = 1
)

type Core struct {
	arch.BaseCore
}

var _ archmod.IActions = &Core{}

func (c *Core) TicksPerBlock() uint {
	return 8
}

func (c *Core) GetMeta() *datamod.MetaRow {
	return datamod.NewMeta(c.Datastore()).Get()
}

func (c *Core) GetBody(bodyId uint8) *datamod.BodiesRow {
	return datamod.NewBodies(c.Datastore()).Get(bodyId)
}

func (c *Core) AddBody(action *archmod.ActionData_AddBody) error {
	meta := c.GetMeta()
	bodyCount := meta.GetBodyCount()
	if bodyCount == math.MaxUint8 {
		return errors.New("too many players")
	}

	bodyId := bodyCount + 1
	meta.SetBodyCount(bodyId)
	body := c.GetBody(bodyId)
	body.Set(action.X, action.Y, action.M, action.Vx, action.Vy)

	return nil
}

func (c *Core) Tick() {
	meta := c.GetMeta()
	bodyCount := meta.GetBodyCount()

	for i := uint8(1); i <= bodyCount; i++ {
		iBody := c.GetBody(i)
		ix, iy := iBody.GetX(), iBody.GetY()
		im := int32(iBody.GetM())
		var ax, ay int32 // Using int32 for possibly larger values during summation

		// Calculate acceleration
		for j := uint8(1); j <= bodyCount; j++ {
			if j == i {
				continue
			}

			jBody := c.GetBody(j)
			jx, jy := jBody.GetX(), jBody.GetY()
			jm := int32(jBody.GetM())

			d := int32(Distance(ix, iy, jx, jy))
			if d == 0 {
				continue // Avoid division by zero, or handle error
			}

			// Compute force, note that we adjust calculations to avoid overflow and maintain precision
			f := G_NUMERATOR * im * jm / (d * d) / G_DENOMINATOR

			dx := int32(jx - ix)
			dy := int32(jy - iy)

			// Calculate acceleration components
			ax += f * dx / d / im // Normalize dx and multiply by force to get acceleration component
			ay += f * dy / d / im // Normalize dy and multiply by force to get acceleration component
		}

		// Update velocities
		vx := iBody.GetVx() + INTERVAL_NUMERATOR*int16(ax)/INTERVAL_DENOMINATOR
		vy := iBody.GetVy() + INTERVAL_NUMERATOR*int16(ay)/INTERVAL_DENOMINATOR
		iBody.SetVx(vx)
		iBody.SetVy(vy)
	}

	// Update positionss
	for i := uint8(1); i <= bodyCount; i++ {
		iBody := c.GetBody(i)
		ix := iBody.GetX() + INTERVAL_NUMERATOR*iBody.GetVx()/INTERVAL_DENOMINATOR
		iy := iBody.GetY() + INTERVAL_NUMERATOR*iBody.GetVy()/INTERVAL_DENOMINATOR
		iBody.SetX(ix)
		iBody.SetY(iy)
	}
}

func distance(x1, y1, x2, y2 int) int {
	dx := utils.Abs(x2 - x1)
	dy := utils.Abs(y2 - y1)
	return (960*utils.Max(dx, dy) + 398*utils.Min(dx, dy)) / 1000
}

func Distance[T constraints.Integer](x1, y1, x2, y2 T) T {
	return T(distance(int(x1), int(y1), int(x2), int(y2)))
}
