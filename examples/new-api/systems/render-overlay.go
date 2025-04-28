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
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/draw"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
)

func NewRenderOverlaySystem() RenderOverlaySystem {
	return RenderOverlaySystem{}
}

type RenderOverlaySystem struct {
	EntityManager                      *ecs.EntityManager
	SceneManager                       *components.AsteroidSceneManagerComponentManager
	Cameras                            *stdcomponents.CameraComponentManager
	FrameBuffer2D                      *stdcomponents.FrameBuffer2DComponentManager
	CollisionChunks                    *stdcomponents.CollisionChunkComponentManager
	CollisionCells                     *stdcomponents.CollisionCellComponentManager
	Tints                              *stdcomponents.TintComponentManager
	BvhTrees                           *stdcomponents.BvhTreeComponentManager
	Positions                          *stdcomponents.PositionComponentManager
	AABBs                              *stdcomponents.AABBComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	Textures                           *stdcomponents.RLTextureProComponentManager
	frameBuffer                        ecs.Entity
	monitorWidth                       int
	monitorHeight                      int
	debugLvl                           int
	debug                              bool
}

func (s *RenderOverlaySystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()

	s.frameBuffer = s.EntityManager.Create()
	s.FrameBuffer2D.Create(s.frameBuffer, stdcomponents.FrameBuffer2D{
		Frame:     rl.Rectangle{X: 0, Y: 0, Width: float32(s.monitorWidth), Height: float32(s.monitorHeight)},
		Texture:   draw.CreateRenderTexture(int32(s.monitorWidth), int32(s.monitorHeight)),
		Layer:     config.DebugLayer,
		BlendMode: rl.BlendAlpha,
		Tint:      rl.White,
		Dst:       rl.Rectangle{Width: float32(s.monitorWidth), Height: float32(s.monitorHeight)},
	})
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

	s.Cameras.EachEntity()(func(entity ecs.Entity) bool {
		camera := s.Cameras.GetUnsafe(entity)
		frame := s.FrameBuffer2D.GetUnsafe(entity)
		switch frame.Layer {
		case config.MainCameraLayer:
			overlayFrame := s.FrameBuffer2D.GetUnsafe(s.frameBuffer)
			draw.BeginTextureMode(overlayFrame.Texture)
			draw.ClearBackground(rl.Blank)

			// Debug mode: BVH tree and dots
			if s.debug {
				draw.BeginMode2D(camera.Camera2D)

				cameraRect := camera.Rect()
				s.CollisionCells.EachEntity()(func(e ecs.Entity) bool {
					cell := s.CollisionCells.GetUnsafe(e)
					assert.NotNil(cell)

					if cell.Layer != stdcomponents.CollisionLayer(s.debugLvl) {
						return true
					}

					tint := s.Tints.GetUnsafe(e)
					assert.NotNil(tint)

					position := s.Positions.GetUnsafe(e)
					assert.NotNil(position)

					tree := s.BvhTrees.GetUnsafe(e)
					assert.NotNil(tree)

					tree.AabbNodes.EachData()(func(a *stdcomponents.AABB) bool {
						// Simple AABB culling
						if s.intersects(cameraRect, a.Rect()) {
							draw.RectFill(
								a.Min.X,
								a.Min.Y,
								a.Max.X-a.Min.X,
								a.Max.Y-a.Min.Y,
								*tint,
							)
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
						Width:  cell.Size,
						Height: cell.Size,
					}) {
						draw.RectLine(position.XY.X, position.XY.Y, cell.Size, cell.Size, 1.0, clr)
					}
					return true
				})
				s.CollisionChunks.EachEntity()(func(e ecs.Entity) bool {
					chunk := s.CollisionChunks.GetUnsafe(e)
					assert.NotNil(chunk)

					if chunk.Layer != stdcomponents.CollisionLayer(s.debugLvl) {
						return true
					}

					tint := s.Tints.GetUnsafe(e)
					assert.NotNil(tint)

					position := s.Positions.GetUnsafe(e)
					assert.NotNil(position)

					tree := s.BvhTrees.GetUnsafe(e)
					assert.NotNil(tree)

					tree.AabbNodes.EachData()(func(a *stdcomponents.AABB) bool {
						// Simple AABB culling
						if s.intersects(cameraRect, a.Rect()) {
							draw.RectFill(
								a.Min.X,
								a.Min.Y,
								a.Max.X-a.Min.X,
								a.Max.Y-a.Min.Y,
								*tint,
							)
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
						draw.RectLine(position.XY.X, position.XY.Y, chunk.Size, chunk.Size, 1.0, clr)
					}
					return true
				})
				s.AABBs.EachEntity()(func(e ecs.Entity) bool {
					aabb := s.AABBs.GetUnsafe(e)
					clr := rl.Green
					isSleeping := s.ColliderSleepStateComponentManager.GetUnsafe(e)
					if isSleeping != nil {
						clr = rl.Blue
					}
					if s.intersects(cameraRect, aabb.Rect()) {
						draw.RectLine(
							aabb.Min.X,
							aabb.Min.Y,
							aabb.Max.X-aabb.Min.X,
							aabb.Max.Y-aabb.Min.Y,
							1.0, clr,
						)
					}
					return true
				})
				s.Collisions.EachEntity()(func(entity ecs.Entity) bool {
					pos := s.Positions.GetUnsafe(entity)
					draw.RectFill(pos.XY.X-8, pos.XY.Y-8, 16, 16, rl.Red)
					return true
				})
				s.Textures.EachComponent()(func(r *stdcomponents.RLTexturePro) bool {
					draw.RectFillAngled(
						r.Dest.X-2,
						r.Dest.Y-2,
						4,
						4,
						r.Rotation, rl.Red,
					)
					return true
				})
				draw.EndMode2D()
			}

			// Print stats
			draw.RectFill(0, 0, 120, 200, rl.Black)
			draw.Text(fmt.Sprintf("FPS: %d", rl.GetFPS()), 10, 10, 20, 2, rl.RayWhite)
			draw.Text(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 70, 20, 2, rl.RayWhite)
			draw.Text(fmt.Sprintf("%d debugLvl", s.debugLvl), 10, 90, 20, 2, rl.RayWhite)
			// Game over
			s.SceneManager.EachComponent()(func(a *components.AsteroidSceneManager) bool {
				draw.Text(fmt.Sprintf("Player HP: %d", a.PlayerHp), 10, 30, 20, 2, rl.RayWhite)
				draw.Text(fmt.Sprintf("Score: %d", a.PlayerScore), 10, 50, 20, 2, rl.RayWhite)
				if a.PlayerHp <= 0 {
					text := "Game Over"
					textSize := rl.MeasureTextEx(rl.GetFontDefault(), text, 96, 2)
					x := (float32(s.monitorWidth) - textSize.X) * 0.5
					y := (float32(s.monitorHeight) - textSize.Y) * 0.5
					draw.Text(text, x, y, 96, 2, rl.Red)
				}
				return false
			})
			draw.EndTextureMode()

		case config.MinimapCameraLayer:
			draw.BeginTextureMode(frame.Texture)
			draw.RectLine(1, 1, float32(frame.Texture.Texture.Width-1), float32(frame.Texture.Texture.Height-1), 1.0, rl.Green)
			draw.EndTextureMode()
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
