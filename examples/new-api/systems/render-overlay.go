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
	"github.com/negrel/assert"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"image/color"
	"time"
)

func NewRenderOverlaySystem() RenderOverlaySystem {
	return RenderOverlaySystem{}
}

type RenderOverlaySystem struct {
	EntityManager   *ecs.EntityManager
	SceneManager    *components.AsteroidSceneManagerComponentManager
	Cameras         *stdcomponents.CameraComponentManager
	FrameBuffer2D   *stdcomponents.FrameBuffer2DComponentManager
	CollisionChunks *stdcomponents.CollisionChunkComponentManager
	Tints           *stdcomponents.TintComponentManager
	BvhTrees        *stdcomponents.BvhTreeComponentManager
	Positions       *stdcomponents.PositionComponentManager
	monitorWidth    int
	monitorHeight   int
	debugLvl        int
	debug           bool
}

func (s *RenderOverlaySystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
}

func (s *RenderOverlaySystem) Run(dt time.Duration) bool {
	if rl.IsKeyPressed(rl.KeyF6) {
		if !s.debug {
			s.debug = true
		} else {
			s.debug = false
		}
	}
	if rl.IsKeyPressed(rl.KeyF7) {
		s.debugLvl--
		if s.debugLvl < 0 {
			s.debugLvl = 63
		}
	}
	if rl.IsKeyPressed(rl.KeyF8) {
		s.debugLvl++
		if s.debugLvl > 63 {
			s.debugLvl = 0
		}
	}
	s.FrameBuffer2D.EachEntity(func(entity ecs.Entity) bool {
		frame := s.FrameBuffer2D.Get(entity)
		camera := s.Cameras.Get(entity)
		switch frame.Layer {
		case config.MainCameraLayer:
			rl.BeginTextureMode(frame.Texture)

			// BVH tree
			if s.debug {
				rl.BeginMode2D(camera.Camera2D)
				cameraRect := camera.Rect()
				s.CollisionChunks.EachEntity(func(e ecs.Entity) bool {
					chunk := s.CollisionChunks.Get(e)
					assert.NotNil(chunk)

					if chunk.Layer != stdcomponents.CollisionLayer(s.debugLvl) {
						return true
					}

					tint := s.Tints.Get(e)
					assert.NotNil(tint)

					position := s.Positions.Get(e)
					assert.NotNil(position)

					tree := s.BvhTrees.Get(e)
					assert.NotNil(tree)

					tree.AabbNodes.EachData(func(a *stdcomponents.AABB) bool {
						// Simple AABB culling
						if s.intersects(cameraRect, a.Rect()) {
							rl.DrawRectangle(int32(a.Min.X), int32(a.Min.Y), int32(a.Max.X-a.Min.X), int32(a.Max.Y-a.Min.Y), *tint)
						}
						return true
					})

					clr := color.RGBA{
						R: tint.R,
						G: tint.G,
						B: tint.B,
						A: 255,
					}

					// Simple AABB culling
					if s.intersects(cameraRect, vectors.Rectangle{
						X:      position.XY.X,
						Y:      position.XY.Y,
						Width:  chunk.Size,
						Height: chunk.Size,
					}) {
						rl.DrawRectangleLines(int32(position.XY.X), int32(position.XY.Y), int32(chunk.Size), int32(chunk.Size), clr)
					}
					return true
				})
				rl.EndMode2D()
			}

			// Print stats
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
			rl.BeginTextureMode(frame.Texture)
			rl.DrawRectangleLines(1, 1, frame.Texture.Texture.Width-1, frame.Texture.Texture.Height-1, rl.Green)
			rl.EndTextureMode()
		}
		return true
	})

	return true
}

func (s *RenderOverlaySystem) intersects(rect1, rect2 vectors.Rectangle) bool {
	return rect1.X < rect2.X+rect2.Width &&
		rect1.X+rect1.Width > rect2.X &&
		rect1.Y < rect2.Y+rect2.Height &&
		rect1.Y+rect1.Height > rect2.Y
}

func (s *RenderOverlaySystem) Destroy() {}
