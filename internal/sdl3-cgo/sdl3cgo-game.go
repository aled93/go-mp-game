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
import "github.com/jfreymuth/go-sdl3/sdl"

func main() {
	must(sdl.Init(sdl.InitVideo))
	defer sdl.Quit()

	w, r, e := sdl.CreateWindowAndRenderer("sld c-go", 640, 480, sdl.WindowResizable)
	must(e)
	defer w.Destroy()
	defer r.Destroy()

Outer:
	for {
		var event sdl.Event
		sdl.PollEvent(&event)
		switch event.Type() {
		case sdl.EventQuit:
			break Outer
		case sdl.EventKeyDown:
			if event.Keyboard().Scancode == sdl.ScancodeEscape {
				break Outer
			}
		}

		must(r.SetDrawColor(0, 0, 0, 255))
		must(r.Clear())
		must(r.SetDrawColor(255, 255, 255, 255))
		must(r.DebugText(10, 10, "Hello"))
		must(r.Present())
	}
}

func must(err error) {
	if err != nil {
		println(err.Error())
		panic(err)
	}
}
