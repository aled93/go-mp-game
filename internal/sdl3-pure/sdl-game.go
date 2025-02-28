package main

import (
	"fmt"
	"github.com/jupiterrider/purego-sdl3/sdl"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	f, err := os.Create("cpu.out")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	fmt.Println("CPU Profile Started")
	defer fmt.Println("CPU Profile Stopped")

	//if !sdl3-pure.SetHint(sdl3-pure.HintFramebufferAcceleration, "1") {
	//	panic(sdl3-pure.GetError())
	//}

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	var window *sdl.Window
	var renderer *sdl.Renderer

	if !sdl.CreateWindowAndRenderer("Hello, World!", 1280, 720, sdl.WindowResizable, &window, &renderer) {
		panic(sdl.GetError())
	}
	defer sdl.DestroyRenderer(renderer)
	defer sdl.DestroyWindow(window)

	sdl.SetRenderDrawColor(renderer, 100, 150, 200, 255)
	var dt time.Duration = time.Second
	var fps int64
	updatedAt := time.Now()
Outer:
	for {
		fps = int64(1 / dt.Seconds())
		dt = time.Since(updatedAt)
		updatedAt = time.Now()

		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				break Outer
			case sdl.EventKeyDown:
				if event.Key().Scancode == sdl.ScancodeEscape {
					break Outer
				}
			}
		}
		sdl.SetRenderDrawColor(renderer, 0, 0, 0, 255)
		sdl.RenderClear(renderer)

		sdl.SetRenderDrawColor(renderer, 255, 255, 255, 255)
		rect := sdl.FRect{300, 300, 200, 200}
		for range 100_000 {
			sdl.RenderFillRect(renderer, &rect)
		}

		sdl.RenderDebugText(renderer, 10, 10, fmt.Sprintf("FPS %d", fps))
		sdl.RenderDebugText(renderer, 10, 20, fmt.Sprintf("dt %s", dt.String()))

		sdl.RenderPresent(renderer)

	}
}
