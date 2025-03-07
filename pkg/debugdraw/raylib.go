package debugdraw

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Render(cam rl.Camera2D) {
	rl.BeginMode2D(cam)
	defer rl.EndMode2D()

	for _, e := range nonfixedDrawEntries {
		renderDrawEntry(&e)
	}
	for _, e := range fixedDrawEntries {
		renderDrawEntry(&e)
	}
}

func renderDrawEntry(e *drawEntry) {
	switch e.Kind {
	case drawKind_Line:
		rl.DrawLine(int32(e.Values[0]), int32(e.Values[1]), int32(e.Values[3]), int32(e.Values[4]), color.RGBA{
			R: uint8(e.R * 255.0),
			G: uint8(e.G * 255.0),
			B: uint8(e.B * 255.0),
			A: uint8(e.A * 255.0),
		})
	}
}
