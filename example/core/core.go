package core

import (
	"errors"
	"math"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/example/gogen/datamod"
)

const (
	Direction_Up uint8 = iota
	Direction_Down
	Direction_Left
	Direction_Right
)

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

var _ archmod.IActions = &Core{}

func (c *Core) GetMeta() *datamod.MetaRow {
	return datamod.NewMeta(c.Datastore()).Get()
}

func (c *Core) GetPlayer(playerId uint16) *datamod.PlayersRow {
	return datamod.NewPlayers(c.Datastore()).Get(playerId)
}

func (c *Core) GetBoardPosition(x, y int16) *datamod.BoardRow {
	return datamod.NewBoard(c.Datastore()).Get(x, y)
}

func (c *Core) AddPlayer(action *archmod.ActionData_AddPlayer) error {
	meta := c.GetMeta()
	playerCount := meta.GetPlayerCount()
	if playerCount == math.MaxUint16 {
		return errors.New("too many players")
	}

	boardPosition := c.GetBoardPosition(action.X, action.Y)
	if boardPosition.GetPlayerId() != 0 {
		return errors.New("position not empty")
	}

	playerId := playerCount + 1

	meta.SetPlayerCount(playerId)
	boardPosition.SetPlayerId(playerId)

	player := c.GetPlayer(playerId)
	player.SetX(action.GetX())
	player.SetY(action.GetY())

	return nil
}

func (c *Core) Move(action *archmod.ActionData_Move) error {
	meta := c.GetMeta()
	playerCount := meta.GetPlayerCount()
	if action.PlayerId > playerCount {
		return errors.New("player does not exist")
	}

	player := c.GetPlayer(action.PlayerId)
	x, y := player.GetX(), player.GetY()

	dX, dY := 0, 0

	switch action.Direction {
	case Direction_Up:
		dY = -1
	case Direction_Down:
		dY = 1
	case Direction_Left:
		dX = -1
	case Direction_Right:
		dX = 1
	}

	newX, ok := safeAddInt16(x, int16(dX))
	if !ok {
		return errors.New("over/underflow")
	}

	newY, ok := safeAddInt16(y, int16(dY))
	if !ok {
		return errors.New("over/underflow")
	}

	newBoardPosition := c.GetBoardPosition(newX, newY)
	if newBoardPosition.GetPlayerId() != 0 {
		return errors.New("position not empty")
	}

	boardPosition := c.GetBoardPosition(x, y)
	boardPosition.SetPlayerId(0)

	newBoardPosition.SetPlayerId(action.PlayerId)
	player.SetX(newX)
	player.SetY(newY)

	return nil
}

func (c *Core) Tick() {}
