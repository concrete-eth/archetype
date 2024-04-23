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
	G_NUMERATOR          int32 = 200
	G_DENOMINATOR        int32 = 1
	INTERVAL_NUMERATOR   int32 = 1
	INTERVAL_DENOMINATOR int32 = 1
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

// TODO: should this be a method [?]
func (c *Core) Mass(r int32) int32 {
	return r * r
}

func (c *Core) NextPosition(body *datamod.BodiesRow) (int32, int32) {
	x := body.GetX() + IntervalDisplacement(body.GetVx(), body.GetAx())
	y := body.GetY() + IntervalDisplacement(body.GetVy(), body.GetAy())
	return x, y
}

func (c *Core) Acceleration(bodyId uint8, body *datamod.BodiesRow) (int32, int32) {
	x, y := body.GetX(), body.GetY()
	r := int32(body.GetR())
	m := c.Mass(r)

	var ax, ay int32

	// Calculate acceleration
	for j := uint8(1); j <= c.GetMeta().GetBodyCount(); j++ {
		if j == bodyId {
			continue
		}

		jBody := c.GetBody(j)
		jx, jy := jBody.GetX(), jBody.GetY()
		jr := int32(jBody.GetR())
		jm := c.Mass(jr)

		d := Distance(x, y, jx, jy)
		if d == 0 || d < r+jr {
			continue
		}

		// Compute force, note that we adjust calculations to avoid overflow and maintain precision
		f := G_NUMERATOR * m * jm / (d * d) / G_DENOMINATOR

		dx := jx - x
		dy := jy - y

		// Calculate acceleration components
		ax += f * dx / d / m // Normalize dx and multiply by force to get acceleration component
		ay += f * dy / d / m // Normalize dy and multiply by force to get acceleration component
	}

	return ax, ay
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
	body.Set(action.X, action.Y, action.R, action.Vx, action.Vy, 0, 0)

	return nil
}

func (c *Core) Tick() {
	meta := c.GetMeta()
	bodyCount := meta.GetBodyCount()

	// Update positionss
	for i := uint8(1); i <= bodyCount; i++ {
		iBody := c.GetBody(i)
		ix, iy := c.NextPosition(iBody)
		iBody.SetX(ix)
		iBody.SetY(iy)
	}

	// Update velocities
	for i := uint8(1); i <= bodyCount; i++ {
		iBody := c.GetBody(i)
		ax, ay := c.Acceleration(i, iBody)

		// Update accelerations
		iBody.SetAx(ax)
		iBody.SetAy(ay)

		// Update velocities
		vx := iBody.GetVx() + INTERVAL_NUMERATOR*ax/INTERVAL_DENOMINATOR
		vy := iBody.GetVy() + INTERVAL_NUMERATOR*ay/INTERVAL_DENOMINATOR
		iBody.SetVx(vx)
		iBody.SetVy(vy)
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

func IntervalDisplacement(v, a int32) int32 {
	return INTERVAL_NUMERATOR*v/INTERVAL_DENOMINATOR +
		a*INTERVAL_NUMERATOR*INTERVAL_NUMERATOR/(2*INTERVAL_DENOMINATOR*INTERVAL_DENOMINATOR)
}
