/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- тажефигня Donated 500 RUB

Thank you for your support!
*/

package core

import (
	"gomp/pkg/draw"
	"gomp/pkg/worker"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	MaxFrameSkips = 5
)

func NewEngine(game AnyGame) Engine {
	numCpu := max(runtime.NumCPU(), 1)
	engine := Engine{
		Game:         game,
		pool:         worker.NewPool(numCpu),
		swapBufferMx: new(sync.Mutex),
	}
	return engine
}

type Engine struct {
	Game         AnyGame
	pool         worker.Pool
	swapBufferMx *sync.Mutex
}

func (e *Engine) Run(tickrate uint, framerate uint) {
	fixedUpdDuration := time.Second / time.Duration(tickrate)

	var renderTicker *time.Ticker
	if framerate > 0 {
		renderTicker = time.NewTicker(time.Second / time.Duration(framerate))
		defer renderTicker.Stop()
	}

	jobChan := make(chan draw.Job, 1)
	exitChan := make(chan struct{})

	draw.SetJobProcessor(jobChan)

	go func() {
		e.Game.Init(e)

		var lastUpdateAt = time.Now() // TODO: REMOVE?
		var nextFixedUpdateAt = time.Now()
		var dt = time.Since(lastUpdateAt)

		e.pool.Start()

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
				nextFixedUpdateAt = time.Now()
				log.Println("Too many updates detected")
			}

			draw.BeginDrawing()
			e.Game.Render(dt)
			draw.EndDrawing()
		}

		e.Game.Destroy()

		draw.SetJobProcessor(nil)
		exitChan <- struct{}{}
	}()

loop:
	for {
		select {
		case job := <-jobChan:
			job.Execute()

		case <-exitChan:
			break loop
		}
	}
}

func (e *Engine) Pool() *worker.Pool {
	return &e.pool
}
