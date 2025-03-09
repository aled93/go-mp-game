/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- HromRu Donated 2500 RUB
<- HromRu Donated 5000 RUB

Thank you for your support!
*/

package gomp

import (
	"log"
	"time"
)

const (
	MaxFrameSkips = 50
)

func NewEngine(game AnyGame) Engine {
	engine := Engine{
		Game: game,
	}

	return engine
}

type Engine struct {
	Game AnyGame
}

func (e *Engine) Run(tickrate uint, framerate uint) {

	fixedUpdDuration := time.Second / time.Duration(tickrate)

	var renderTicker *time.Ticker
	if framerate > 0 {
		renderTicker = time.NewTicker(time.Second / time.Duration(framerate))
		defer renderTicker.Stop()
	}

	e.Game.Init()
	defer e.Game.Destroy()

	var lastUpdateAt = time.Now() // TODO: REMOVE?
	var nextFixedUpdateAt = time.Now()
	var dt = time.Since(lastUpdateAt)

	for !e.Game.ShouldDestroy() {
		if renderTicker != nil {
			<-renderTicker.C
		}
		dt = time.Since(lastUpdateAt)
		lastUpdateAt = time.Now()

		// Update
		e.Game.Update(dt)

		// Fixed Update
		loops := 0
		// TODO: Refactor to work without for loop
		for nextFixedUpdateAt.Compare(time.Now()) == -1 && loops < MaxFrameSkips {
			e.Game.FixedUpdate(fixedUpdDuration)
			nextFixedUpdateAt = nextFixedUpdateAt.Add(fixedUpdDuration)
			loops++
		}
		if loops >= MaxFrameSkips {
			log.Println("Too many updates detected")
		}

		// RenderAssterodd
		e.Game.Render(dt)
	}
}
