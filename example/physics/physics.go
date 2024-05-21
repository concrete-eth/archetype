package physics

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
	SCALE          = 100
	GRAV     int32 = 40 * SCALE
	INTERVAL int32 = 0.25 * SCALE
)

// mul multiplies multiple int32 fixed-point numbers and caps the result at int32 limits.
func mul(m ...int32) int32 {
	product := int32(SCALE)
	for _, num := range m {
		if num == 0 {
			return 0
		}
		if utils.Abs(product) > math.MaxInt32/utils.Abs(num) {
			if utils.Abs(product) > math.MaxInt32/utils.Abs(num)/SCALE {
				panic("overflow")
			} else {
				product = product / SCALE * num
				continue
			}
		}
		product = product * num / SCALE
	}
	return product
}

// div divides two int32 fixed-point numbers and adjusts for overflows.
func div(a, b int32) int32 {
	if b == 0 {
		panic("division by zero")
	}
	if a > math.MaxInt32/SCALE || a < math.MinInt32/SCALE {
		panic("overflow")
	}
	return a * SCALE / b
}

type Core struct {
	arch.BaseCore
}

var _ archmod.IActions = &Core{}

func (c *Core) TicksPerBlock() uint64 {
	return 8
}

func (c *Core) GetMeta() *datamod.MetaRow {
	return datamod.NewMeta(c.Datastore()).Get()
}

func (c *Core) GetBody(bodyId uint8) *datamod.BodiesRow {
	return datamod.NewBodies(c.Datastore()).Get(bodyId)
}

func (c *Core) NextPosition(body *datamod.BodiesRow) (int32, int32) {
	x := body.GetX() + IntervalDisplacement(body.GetVx(), body.GetAx())
	y := body.GetY() + IntervalDisplacement(body.GetVy(), body.GetAy())
	return x, y
}

func (c *Core) Mass(r int32) int32 {
	return mul(r, r)
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
		if d == 0 || d < r+jr || d > math.MaxInt16 {
			continue
		}

		// Compute force, note that we adjust calculations to avoid overflow and maintain precision
		f := div(mul(GRAV, m, jm), mul(d, d))

		dx := jx - x
		dy := jy - y
		dxy := utils.Abs(dx) + utils.Abs(dy)

		// Calculate acceleration components
		ax += div(mul(f, dx), mul(dxy, m)) // Normalize dx and multiply by force to get acceleration component
		ay += div(mul(f, dy), mul(dxy, m)) // Normalize dy and multiply by force to get acceleration component
	}

	return ax, ay
}

func (c *Core) AddBody(action *archmod.ActionData_AddBody) error {
	if utils.Abs(action.X) > math.MaxInt16 || utils.Abs(action.Y) > math.MaxInt16 {
		return errors.New("position out of bounds")
	}
	if action.R == 0 {
		return errors.New("radius too small")
	}
	if action.R > 100*SCALE {
		return errors.New("radius too large")
	}
	if utils.Abs(action.Vx) > 100*SCALE || utils.Abs(action.Vy) > 100*SCALE {
		return errors.New("velocity too large")
	}

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
		vx := iBody.GetVx() + mul(INTERVAL, ax)
		vy := iBody.GetVy() + mul(INTERVAL, ay)
		iBody.SetVx(vx)
		iBody.SetVy(vy)
	}
}

// Distance calculates the distance between two points.
func Distance[T constraints.Integer](x1, y1, x2, y2 T) T {
	return T(distance(int(x1), int(y1), int(x2), int(y2)))
}

func distance(x1, y1, x2, y2 int) int {
	dx := utils.Abs(x2 - x1)
	dy := utils.Abs(y2 - y1)
	return (960*utils.Max(dx, dy) + 398*utils.Min(dx, dy)) / 1000
}

// IntervalDisplacement calculates the displacement of an object over an interval.
func IntervalDisplacement(v, a int32) int32 {
	return mul(INTERVAL, v) + div(mul(a, INTERVAL, INTERVAL), 2*SCALE)
}
