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

package systems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewRenderOverlaySystem() RenderOverlaySystem {
	return RenderOverlaySystem{}
}

type RenderOverlaySystem struct {
	EntityManager *ecs.EntityManager
	SceneManager  *components.AsteroidSceneManagerComponentManager
	Cameras       *stdcomponents.CameraComponentManager
	FrameBuffer2D *stdcomponents.FrameBuffer2DComponentManager
	monitorWidth  int
	monitorHeight int
}

func (s *RenderOverlaySystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
}

func (s *RenderOverlaySystem) Run(dt time.Duration) bool {

	s.FrameBuffer2D.EachComponent(func(c *stdcomponents.FrameBuffer2D) bool {
		switch c.Layer {
		case config.MainCameraLayer:
			rl.BeginTextureMode(c.Texture)
			// Print the current FPS
			rl.DrawRectangleRec(rl.Rectangle{Height: 100, Width: 200}, rl.Black)
			rl.DrawFPS(10, 10)
			rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)
			// Game over
			s.SceneManager.EachComponent(func(a *components.AsteroidSceneManager) bool {
				rl.DrawText(fmt.Sprintf("Player HP: %d", a.PlayerHp), 10, 50, 20, rl.RayWhite)
				rl.DrawText(fmt.Sprintf("Score: %d", a.PlayerScore), 10, 70, 20, rl.RayWhite)
				if a.PlayerHp <= 0 {
					text := "Game Over"
					textSize := rl.MeasureTextEx(rl.GetFontDefault(), text, 96, 0)
					x := (s.monitorWidth - int(textSize.X)) / 2
					y := (s.monitorHeight - int(textSize.Y)) / 2
					rl.DrawText(text, int32(x), int32(y), 96, rl.Red)

				}
				return false
			})
			rl.EndTextureMode()
		case config.MinimapCameraLayer:
			rl.BeginTextureMode(c.Texture)
			rl.DrawRectangleLines(1, 1, c.Texture.Texture.Width-1, c.Texture.Texture.Height-1, rl.Green)
			rl.EndTextureMode()
		}
		return true
	})

	return true
}

func (s *RenderOverlaySystem) Destroy() {}
