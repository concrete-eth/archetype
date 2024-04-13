package main

import (
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/example/client"
	"github.com/concrete-eth/archetype/example/gogen/archmod"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	var (
		kv                                     = kvstore.NewMemoryKeyValueStore()
		actionBatchInChan                      = make(chan arch.ActionBatch, 1)
		actionOutChan     chan<- []arch.Action = nil
		blockTime                              = 1 * time.Second
		blockNumber       uint64               = 0
	)

	go func() {
		bn := blockNumber
		actionBatchInChan <- arch.ActionBatch{
			BlockNumber: bn,
			Actions: []arch.Action{
				&arch.CanonicalTickAction{},
				&archmod.ActionData_AddBody{X: 0, Y: 0, M: 1000, Vx: 0, Vy: 0},
				&archmod.ActionData_AddBody{X: -500, Y: 0, M: 250, Vx: 0, Vy: -20},
				// &archmod.ActionData_AddBody{X: 500, Y: 0, M: 250, Vx: 0, Vy: 10},
			},
		}
		ticker := time.NewTicker(blockTime)
		for range ticker.C {
			bn++
			actionBatchInChan <- arch.ActionBatch{
				BlockNumber: bn,
				Actions: []arch.Action{
					&arch.CanonicalTickAction{},
				},
			}
		}
	}()

	c := client.NewClient(kv, actionBatchInChan, actionOutChan, blockTime, blockNumber)

	w, h := c.Layout(0, 0)

	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Archetype Example")
	ebiten.SetTPS(30)

	if err := ebiten.RunGame(c); err != nil {
		panic(err)
	}
}
