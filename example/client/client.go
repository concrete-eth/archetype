package client

import (
	"image/color"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/client"
	"github.com/concrete-eth/archetype/example/core"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/gogen/datamod"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	TrailLength = uint64(16)
)

type Client struct {
	client.Client
	positionHistory map[uint8]map[uint64][2]int16
	width           int
	height          int
}

func NewClient(
	kv lib.KeyValueStore,
	actionBatchInChan <-chan arch.ActionBatch,
	actionOutChan chan<- []arch.Action,
	blockTime time.Duration,
	blockNumber uint64,
) *Client {
	specs := arch.ArchSpecs{Actions: archmod.ActionSpecs, Tables: archmod.TableSpecs}
	c := &core.Core{}
	return &Client{
		Client:          *client.New(specs, c, kv, actionBatchInChan, actionOutChan, blockTime, blockNumber),
		positionHistory: make(map[uint8]map[uint64][2]int16),
		width:           640,
		height:          720,
	}
}

func (c *Client) tickIndex() uint64 {
	core := c.Core()
	return core.BlockNumber()*uint64(core.TicksPerBlock()) + uint64(core.InBlockTickIndex())
}

func (c *Client) internalPositionToScreenPosition(screen *ebiten.Image, x, y int16) (float32, float32) {
	screenX := float32(x+1000) / 2000 * float32(c.width)
	screenY := float32(y+1000) / 2000 * float32(c.height)
	return screenX, screenY
}

func (c *Client) GetBodyCount() uint8 {
	return c.Core().(*core.Core).GetMeta().GetBodyCount()
}

func (c *Client) GetBody(bodyId uint8) *datamod.BodiesRow {
	return c.Core().(*core.Core).GetBody(bodyId)
}

func (c *Client) Update() error {
	_, didTick, err := c.InterpolatedSync()
	if err != nil {
		return err
	}
	if didTick {
		tickIndex := c.tickIndex()
		bodyCount := c.GetBodyCount()
		for i := uint8(1); i <= bodyCount; i++ {
			body := c.GetBody(i)
			x, y := body.GetX(), body.GetY()
			if c.positionHistory[i] == nil {
				c.positionHistory[i] = make(map[uint64][2]int16)
			}
			c.positionHistory[i][tickIndex] = [2]int16{x, y}
			delete(c.positionHistory[i], tickIndex-TrailLength-1)
		}
	}
	return nil
}

func (c *Client) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	bodyCount := c.Core().(*core.Core).GetMeta().GetBodyCount()

	// Draw trails
	for i := uint8(1); i <= bodyCount; i++ {
		if c.positionHistory[i] == nil {
			continue
		}

		curIndex := c.tickIndex()

		// Find the start of the trail
		var trailStart uint64
		if TrailLength > curIndex {
			trailStart = 0
		} else {
			trailStart = curIndex - TrailLength
		}
		var lastPos [2]int16
		var lastPosOk bool
		for idx := trailStart; idx < curIndex; idx++ {
			if pos, ok := c.positionHistory[i][idx]; ok {
				trailStart = idx + 1
				lastPos = pos
				lastPosOk = true
				break
			}
		}
		if !lastPosOk {
			continue
		}

		// Draw the trail
		for idx := trailStart; idx <= curIndex; idx++ {
			pos, ok := c.positionHistory[i][idx]
			if !ok {
				continue
			}
			psx, psy := c.internalPositionToScreenPosition(screen, lastPos[0], lastPos[1])
			sx, sy := c.internalPositionToScreenPosition(screen, pos[0], pos[1])
			sr := float32(1)
			vector.StrokeLine(screen, psx, psy, sx, sy, sr, color.White, true)
			lastPos = pos
		}
	}

	// Draw bodies
	for i := uint8(1); i <= bodyCount; i++ {
		body := c.GetBody(i)
		x, y := body.GetX(), body.GetY()
		m := body.GetM()
		sx, sy := c.internalPositionToScreenPosition(screen, x, y)
		sr := float32(m / 100)
		vector.DrawFilledCircle(screen, sx, sy, sr, color.White, true)
	}
}

func (c *Client) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return c.width, c.height
}
