package main

import (
	"context"
	"gomp_game/pkgs/gomp-v0/engine"
	"gomp_game/pkgs/gomp-v0/example/scenes"

	"time"
)

const tickRate = time.Second

func main() {
	e := engine.NewEngine(tickRate).SetDebug(true).SetDebugDraw(false)

	e.LoadScene(scenes.VillageScene)

	e.Run(context.Background())
}
