package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	fontPath = "test.ttf"
	fontSize = 32
)

var winTitle string = "Go-SDL2 Render"
var winWidth, winHeight int32 = 800, 600

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var font *ttf.Font
	var text *sdl.Surface
	var rects []sdl.FRect = make([]sdl.FRect, 100_000)

	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	for i := range len(rects) {
		rects[i] = sdl.FRect{X: float32(i%780) + 10, Y: float32(i / 780), W: 1, H: 1}
	}

	// Load the font for our text
	if font, err = ttf.OpenFont(fontPath, fontSize); err != nil {
		return 0
	}
	defer font.Close()

	// Create a red text with the font
	if text, err = font.RenderUTF8Blended("Hello, World!", sdl.Color{R: 255, G: 0, B: 0, A: 255}); err != nil {
		return 0
	}
	defer text.Free()

	// Draw the text around the center of the window

	var dt time.Duration = time.Second
	var fps int64
	updatedAt := time.Now()

	running := true
	for running {
		fps = int64(1 / dt.Seconds())
		dt = time.Since(updatedAt)
		updatedAt = time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		must(renderer.SetDrawColor(0, 0, 0, 255))
		must(renderer.Clear())

		must(renderer.SetDrawColor(255, 255, 255, 255))
		//rect = sdl.FRect{300, 0, 200, 200}

		//for range 100_000 {
		//	must(renderer.DrawRectF(&rect))
		//}

		//must(text.(nil, renderer, &sdl.Rect{X: 400 - (text.W / 2), Y: 300 - (text.H / 2), W: 0, H: 0}))

		must(renderer.FillRectsF(rects))
		renderer.Present()
		log.Printf("FPS: %d\n", fps)
	}

	return 0
}

func must(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func main() {
	os.Exit(run())
}
