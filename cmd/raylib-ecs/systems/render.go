/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"fmt"
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
	"gomp_game/pkgs/spatial"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type renderController struct {
	width, height int32
	texture       rl.Texture2D
	showDebugInfo bool
	showQtreeVis  bool
}

func (s *renderController) Init(world *ecs.World) {
	rl.InitWindow(s.width, s.height, "raylib [core] example - basic window")

	// currentMonitorRefreshRate := rl.GetMonitorRefreshRate(rl.GetCurrentMonitor())
	// // rl.SetTargetFPS(int32(currentMonitorRefreshRate))

	s.texture = rl.LoadTexture("assets/star.png")
}

func (s *renderController) Update(world *ecs.World) {
	spriteRenders := components.SpriteRenderService.GetManager(world)
	sprites := components.SpriteService.GetManager(world)

	sprites.AllDataParallel(func(sprite *components.Sprite) bool {
		if sprite.Texture == nil {
			sprite.Texture = &s.texture
		}

		sprite.TextureRegion = rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(s.texture.Width),
			Height: float32(s.texture.Height),
		}
		return true
	})

	if rl.WindowShouldClose() {
		world.SetShouldDestroy(true)
		return
	}

	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.Black)

	spriteRenders.AllData(func(spriteRender *components.SpriteRender) bool {
		sprite := &spriteRender.Sprite
		dest := spriteRender.Dest
		texture := *sprite.Texture

		rl.DrawTexturePro(texture, sprite.TextureRegion, dest, sprite.Origin, spriteRender.Rotation, sprite.Tint)
		return true
	})

	if s.showQtreeVis {
		drawUserDataDebug(gravity.QTree.Root())
	}

	if s.showDebugInfo {
		rl.DrawRectangle(0, 0, 120, 100, rl.DarkGray)
		rl.DrawFPS(10, 10)
		rl.DrawText(fmt.Sprintf("%d", world.Size()), 10, 30, 20, rl.Red)
	}

	if rl.IsKeyPressed(rl.KeyF1) {
		s.showDebugInfo = !s.showDebugInfo
	}
	if rl.IsKeyPressed(rl.KeyF2) {
		s.showQtreeVis = !s.showQtreeVis
	}
}

func (s *renderController) FixedUpdate(world *ecs.World) {}

func (s *renderController) Destroy(world *ecs.World) {
	rl.CloseWindow()
}

func drawUserDataDebug(n *spatial.QuadNode[gravity.QuadNodeUserData, any]) {
	clr := rl.Lime
	minx, miny, maxx, maxy := n.Bounds()

	// if ents := n.Entities(); ents != nil {
	// 	rl.DrawText(strconv.Itoa(len(ents)), int32(minx), int32(maxy)-10, 10, clr)
	// }

	if n.Childs()[0] != nil {
		midx := minx + (maxx-minx)*0.5
		midy := miny + (maxy-miny)*0.5

		rl.DrawLine(int32(minx), int32(midy), int32(maxx), int32(midy), clr)
		rl.DrawLine(int32(midx), int32(miny), int32(midx), int32(maxy), clr)

		rl.DrawLine(int32(midx), int32(midy), int32(n.UserData().GX), int32(n.UserData().GY), rl.Red)
		rl.DrawCircle(int32(n.UserData().GX), int32(n.UserData().GY), float32(n.UserData().Mass)*0.001, rl.Red)
		rl.DrawText(fmt.Sprintf("%.1fx%.1f", n.UserData().GX, n.UserData().GY), int32(minx), int32(miny), 10, rl.Lime)

		for _, child := range n.Childs() {
			drawUserDataDebug(child)
		}
	}
}
