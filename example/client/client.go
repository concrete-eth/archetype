package client

import (
	"image/color"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/client"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/gogen/datamod"
	"github.com/concrete-eth/archetype/example/physics"
	"github.com/concrete-eth/archetype/rpc"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	TrailLength  = uint64(16)
	ScreenWidth  = 640
	ScreenHeight = 720
	PixelSize    = 2.0
)

type Client struct {
	*client.Client
	hinter *rpc.TxHinter

	lastTickTime      time.Time
	positionHistory   map[uint8]map[uint64][2]int32 // bodyId -> tickIndex -> [x, y]
	hinterNonce       uint64
	anticipatedBodies map[uint8][3]int32 // bodyId -> [x, y, r]

	_tickTime time.Duration
}

func NewClient(
	kv lib.KeyValueStore,
	io *rpc.IO,
) *Client {
	c := &physics.Core{}
	cli := io.NewClient(kv, c)
	hinter := io.Hinter()
	return &Client{
		Client: cli,
		hinter: hinter,

		lastTickTime:      time.Now(),
		positionHistory:   make(map[uint8]map[uint64][2]int32),
		hinterNonce:       0,
		anticipatedBodies: make(map[uint8][3]int32),

		_tickTime: cli.BlockTime() / time.Duration(c.TicksPerBlock()),
	}
}

func (c *Client) tickIndex() uint64 {
	core := c.Core()
	return core.BlockNumber()*uint64(core.TicksPerBlock()) + uint64(core.InBlockTickIndex())
}

func (c *Client) updatePositionHistory() {
	tickIndex := c.tickIndex()
	bodyCount := c.GetBodyCount()
	for i := uint8(1); i <= bodyCount; i++ {
		body := c.GetBody(i)
		x, y := body.GetX(), body.GetY()
		if c.positionHistory[i] == nil {
			c.positionHistory[i] = make(map[uint64][2]int32)
		}
		c.positionHistory[i][tickIndex] = [2]int32{x, y}
		delete(c.positionHistory[i], tickIndex-TrailLength-1)
	}
}

func (c *Client) coreCoordToScreenCoord(x, y int32) (float32, float32) {
	screenX := float32(x)/PixelSize + ScreenWidth/2
	screenY := float32(y)/PixelSize + ScreenHeight/2
	return screenX, screenY
}

func (c *Client) screenCoordToCoreCoord(x, y float32) (int32, int32) {
	coreX := int32((x - ScreenWidth/2) * PixelSize)
	coreY := int32((y - ScreenHeight/2) * PixelSize)
	return coreX, coreY
}

func (c *Client) Core() *physics.Core {
	return c.Client.Core().(*physics.Core)
}

func (c *Client) AddBody(x, y, r int32) {
	c.SendAction(&archmod.ActionData_AddBody{
		X: x,
		Y: y,
		R: uint32(r),
	})
}

func (c *Client) GetMeta() *datamod.MetaRow {
	return c.Core().GetMeta()
}

func (c *Client) GetBody(bodyId uint8) *datamod.BodiesRow {
	return c.Core().GetBody(bodyId)
}

func (c *Client) GetBodyCount() uint8 {
	return c.GetMeta().GetBodyCount()
}

func (c *Client) Update() error {
	_, didTick, err := c.InterpolatedSync()
	if err != nil {
		return err
	}
	if didTick {
		c.lastTickTime = time.Now()
		c.updatePositionHistory()
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		coreX, coreY := c.screenCoordToCoreCoord(float32(x), float32(y))
		c.AddBody(coreX, coreY, 10)
	}

	if c.hinter != nil && c.hinter.HintNonce() > c.hinterNonce {
		actualBodyCount := c.GetBodyCount()
		c.anticipatedBodies = make(map[uint8][3]int32)
		_, hints := c.hinter.GetHints()

		c.Simulate(func(_core arch.Core) {
			core := _core.(*physics.Core)
			for _, actions := range hints {
				for _, action := range actions {
					switch action := action.(type) {
					case *archmod.ActionData_AddBody:
						core.AddBody(action)
					}
				}
			}
			simBodyCount := c.GetMeta().GetBodyCount()
			for i := uint8(actualBodyCount + 1); i <= simBodyCount; i++ {
				body := c.GetBody(i)
				c.anticipatedBodies[i] = [3]int32{body.GetX(), body.GetY(), int32(body.GetR())}
			}
		})
	}

	return nil
}

func (c *Client) drawLine(screen *ebiten.Image, x1, y1, x2, y2 int32) {
	sx1, sy1 := c.coreCoordToScreenCoord(x1, y1)
	sx2, sy2 := c.coreCoordToScreenCoord(x2, y2)
	vector.StrokeLine(screen, sx1, sy1, sx2, sy2, 1, color.White, true)
}

func (c *Client) drawCircle(screen *ebiten.Image, x, y, r int32, anticipated bool) {
	sx, sy := c.coreCoordToScreenCoord(x, y)
	sr := float32(r) / PixelSize
	var clr color.Color
	if anticipated {
		clr = color.RGBA{0x80, 0x80, 0x80, 0xff}
	} else {
		clr = color.White
	}
	vector.DrawFilledCircle(screen, sx, sy, sr, clr, true)
}

func (c *Client) drawTrail(screen *ebiten.Image, bodyId uint8, body *datamod.BodiesRow) {
	if c.positionHistory[bodyId] == nil {
		return
	}

	curIndex := c.tickIndex()

	// Find the start of the trail
	var trailStart uint64
	if TrailLength > curIndex {
		trailStart = 0
	} else {
		trailStart = curIndex - TrailLength
	}
	var lastPos [2]int32
	var lastPosOk bool
	for idx := trailStart; idx < curIndex; idx++ {
		if pos, ok := c.positionHistory[bodyId][idx]; ok {
			trailStart = idx + 1
			lastPos = pos
			lastPosOk = true
			break
		}
	}
	if !lastPosOk {
		return
	}

	// Draw the trail
	for idx := trailStart; idx <= curIndex; idx++ {
		if pos, ok := c.positionHistory[bodyId][idx]; ok {
			c.drawLine(screen, lastPos[0], lastPos[1], pos[0], pos[1])
			lastPos = pos
		}
	}

	// Draw a line from the last position to the interpolated position
	x, y := c.interpolatedPosition(bodyId, body)
	c.drawLine(screen, lastPos[0], lastPos[1], x, y)
}

func (c *Client) interpolatedPosition(bodyId uint8, body *datamod.BodiesRow) (int32, int32) {
	x, y := body.GetX(), body.GetY()
	nx, ny := c.Core().NextPosition(body)
	tickFraction := time.Since(c.lastTickTime).Seconds() / float64(c._tickTime.Seconds())
	ix := x + int32(float64(nx-x)*tickFraction)
	iy := y + int32(float64(ny-y)*tickFraction)
	return ix, iy
}

func (c *Client) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	bodyCount := c.GetBodyCount()
	for bodyId := uint8(1); bodyId <= bodyCount; bodyId++ {
		body := c.GetBody(bodyId)
		x, y := c.interpolatedPosition(bodyId, body)
		r := int32(body.GetR())
		c.drawTrail(screen, bodyId, body)
		c.drawCircle(screen, x, y, r, false)
	}
	for bodyId, state := range c.anticipatedBodies {
		if bodyId <= bodyCount {
			// Body already exists, don't draw it
			continue
		}
		x, y, _r := state[0], state[1], state[2]
		r := int32(_r)
		c.drawCircle(screen, x, y, r, true)
	}
}

func (c *Client) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
