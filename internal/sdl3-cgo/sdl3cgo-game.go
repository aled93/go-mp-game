/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package main

import "C"
import (
	"fmt"
	"github.com/jfreymuth/go-sdl3/sdl"
	"runtime"
	"time"
)

const batchSize = 1 << 14
const rectCounter = 1 << 18
const framerate = 600

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sdl.SetHint(sdl.HintRenderDriver, "gpu")
	must(sdl.Init(sdl.InitVideo))
	defer sdl.Quit()

	w, r, e := sdl.CreateWindowAndRenderer("sld c-go", 2560, 1440, sdl.WindowResizable)
	must(e)
	defer w.Destroy()
	defer r.Destroy()

	var renderTicker *time.Ticker
	if framerate > 0 {
		renderTicker = time.NewTicker(time.Second / time.Duration(framerate))
		defer renderTicker.Stop()
	}

	var dt time.Duration = time.Second
	var fps float64
	updatedAt := time.Now()

	rects := make([]sdl.FRect, rectCounter)
	for i := range rects {
		rects[i] = sdl.FRect{X: 300, Y: 300, W: 10, H: 10}
	}

Outer:
	for {
		if renderTicker != nil {
			<-renderTicker.C
		}
		fps = (time.Second.Seconds() / dt.Seconds())
		dt = time.Since(updatedAt)
		updatedAt = time.Now()

		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				break Outer
			case sdl.EventKeyDown:
				if event.Keyboard().Scancode == sdl.ScancodeEscape {
					break Outer
				}
			}
		}

		must(r.SetDrawColor(0, 0, 0, 255))
		must(r.Clear())

		must(r.SetDrawColor(255, 255, 255, 255))

		for i := 0; i < len(rects); i += batchSize {
			must(r.FillRects(rects[i : i+batchSize]))
		}

		//must(r.FillRects(rects))

		must(r.DebugText(10, 10, fmt.Sprintf("FPS %f", fps)))
		must(r.DebugText(10, 20, fmt.Sprintf("dt %s", dt.String())))

		must(r.Present())
	}
}

func must(err error) {
	if err != nil {
		println(err.Error())
		panic(err)
	}
}
