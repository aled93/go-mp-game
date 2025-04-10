/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- HromRu Donated 1 500 RUB

Thank you for your support!
*/

package stdsystems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"image/color"
	"math"
	"slices"
	"sync"
	"time"
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{
		instanceData: make([]stdcomponents.RLTexturePro, 0, 8192),
	}
}

type RenderSystem struct {
	EntityManager                      *ecs.EntityManager
	RlTexturePros                      *stdcomponents.RLTextureProComponentManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	AnimationPlayers                   *stdcomponents.AnimationPlayerComponentManager
	Tints                              *stdcomponents.TintComponentManager
	Flips                              *stdcomponents.FlipComponentManager
	Renderables                        *stdcomponents.RenderableComponentManager
	AnimationStates                    *stdcomponents.AnimationStateComponentManager
	Sprites                            *stdcomponents.SpriteComponentManager
	SpriteMatrixes                     *stdcomponents.SpriteMatrixComponentManager
	RenderOrders                       *stdcomponents.RenderOrderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	AABBs                              *stdcomponents.AABBComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTrees                           *stdcomponents.BvhTreeComponentManager
	CollisionChunks                    *stdcomponents.CollisionChunkComponentManager

	renderList   []renderEntry
	instanceData []stdcomponents.RLTexturePro
	camera       rl.Camera2D

	monitorWidth  int
	monitorHeight int

	debug    bool
	debugLvl int
}

type renderEntry struct {
	Entity    ecs.Entity
	TextureId int
	ZIndex    float32
}

func (s *RenderSystem) Init() {
	rl.InitWindow(0, 0, "GOMP")
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
	s.camera = rl.Camera2D{
		Target:   rl.NewVector2(float32(s.monitorWidth/2), float32(s.monitorHeight/2)),
		Offset:   rl.NewVector2(float32(s.monitorWidth/2), float32(s.monitorHeight/2)),
		Rotation: 0,
		Zoom:     1,
	}
}

func (s *RenderSystem) Run(dt time.Duration) bool {
	if rl.WindowShouldClose() {
		return true
	}

	if rl.IsKeyPressed(rl.KeyF12) {
		s.debug = !s.debug
	}
	if rl.IsKeyPressed(rl.KeyF11) {
		s.debugLvl++
		if s.debugLvl > 63 {
			s.debugLvl = 0
		}
	}
	if rl.IsKeyPressed(rl.KeyF10) {
		s.debugLvl--
		if s.debugLvl < 0 {
			s.debugLvl = 63
		}
	}

	fdt := float32(dt.Seconds())

	if rl.IsKeyDown(rl.KeyLeft) {
		s.camera.Target.X -= 1000 * fdt
	}
	if rl.IsKeyDown(rl.KeyRight) {
		s.camera.Target.X += 1000 * fdt
	}
	if rl.IsKeyDown(rl.KeyUp) {
		s.camera.Target.Y -= 1000 * fdt
	}
	if rl.IsKeyDown(rl.KeyDown) {
		s.camera.Target.Y += 1000 * fdt
	}

	s.prepareRender(dt)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	s.render()

	rl.DrawFPS(10, 10)
	rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)
	rl.DrawText(fmt.Sprintf("%d debugLvl", s.debugLvl), 10, 50, 20, rl.RayWhite)
	rl.EndDrawing()

	return false
}

func (s *RenderSystem) Destroy() {
	rl.CloseWindow()
}

type RenderInjector struct {
	EntityManager                      *ecs.EntityManager
	RlTexturePros                      *stdcomponents.RLTextureProComponentManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	AnimationPlayers                   *stdcomponents.AnimationPlayerComponentManager
	Tints                              *stdcomponents.TintComponentManager
	Flips                              *stdcomponents.FlipComponentManager
	Renderables                        *stdcomponents.RenderableComponentManager
	AnimationStates                    *stdcomponents.AnimationStateComponentManager
	Sprites                            *stdcomponents.SpriteComponentManager
	SpriteMatrixes                     *stdcomponents.SpriteMatrixComponentManager
	RenderOrders                       *stdcomponents.RenderOrderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	AABBs                              *stdcomponents.AABBComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTrees                           *stdcomponents.BvhTreeComponentManager
	CollisionChunks                    *stdcomponents.CollisionChunkComponentManager
}

func (s *RenderSystem) InjectWorld(injector *RenderInjector) {
	s.EntityManager = injector.EntityManager
	s.RlTexturePros = injector.RlTexturePros
	s.Positions = injector.Positions
	s.Rotations = injector.Rotations
	s.Scales = injector.Scales
	s.AnimationPlayers = injector.AnimationPlayers
	s.Tints = injector.Tints
	s.Flips = injector.Flips
	s.Renderables = injector.Renderables
	s.AnimationStates = injector.AnimationStates
	s.Sprites = injector.Sprites
	s.SpriteMatrixes = injector.SpriteMatrixes
	s.RenderOrders = injector.RenderOrders
	s.BoxColliders = injector.BoxColliders
	s.CircleColliders = injector.CircleColliders
	s.AABBs = injector.AABBs
	s.Collisions = injector.Collisions
	s.ColliderSleepStateComponentManager = injector.ColliderSleepStateComponentManager
	s.BvhTrees = injector.BvhTrees
	s.CollisionChunks = injector.CollisionChunks
}

func (s *RenderSystem) render() {
	// ==========
	// DEBUG
	// ==========
	if s.debug {
		rl.BeginMode2D(s.camera)
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
				rl.DrawRectangle(int32(a.Min.X), int32(a.Min.Y), int32(a.Max.X-a.Min.X), int32(a.Max.Y-a.Min.Y), *tint)
				return true
			})

			clr := color.RGBA{
				R: tint.R,
				G: tint.G,
				B: tint.B,
				A: 255,
			}

			rl.DrawRectangleLines(int32(position.XY.X), int32(position.XY.Y), int32(chunk.Size), int32(chunk.Size), clr)
			return true
		})
		s.BoxColliders.EachEntity(func(e ecs.Entity) bool {
			col := s.BoxColliders.Get(e)
			scale := s.Scales.Get(e)
			pos := s.Positions.Get(e)
			rot := s.Rotations.Get(e)

			rl.DrawRectanglePro(rl.Rectangle{
				X:      pos.XY.X,
				Y:      pos.XY.Y,
				Width:  col.WH.X * scale.XY.X,
				Height: col.WH.Y * scale.XY.Y,
			}, rl.Vector2{
				X: col.Offset.X * scale.XY.X,
				Y: col.Offset.Y * scale.XY.Y,
			}, float32(rot.Degrees()), rl.DarkGreen)
			return true
		})
		s.CircleColliders.EachEntity(func(e ecs.Entity) bool {
			col := s.CircleColliders.Get(e)
			scale := s.Scales.Get(e)
			pos := s.Positions.Get(e)

			color := rl.DarkGreen
			isSleeping := s.ColliderSleepStateComponentManager.Get(e)
			if isSleeping != nil {
				color = rl.Blue
			}

			posWithOffset := pos.XY.Add(col.Offset.Mul(scale.XY))
			rl.DrawCircle(int32(posWithOffset.X), int32(posWithOffset.Y), col.Radius*scale.XY.X, color)
			return true
		})
		rl.EndMode2D()
	}

	// Extract and sort entities
	if cap(s.renderList) < s.Renderables.Len() {
		s.renderList = append(s.renderList, make([]renderEntry, 0, s.Renderables.Len()-cap(s.renderList))...)
	}
	s.Renderables.EachEntity(func(e ecs.Entity) bool {
		renderOrder := s.RenderOrders.Get(e)

		spriteMatrix := s.SpriteMatrixes.Get(e)
		if spriteMatrix != nil {
			s.renderList = append(s.renderList, renderEntry{
				Entity:    e,
				TextureId: int(spriteMatrix.Texture.ID),
				ZIndex:    renderOrder.CalculatedZ,
			})
			return true
		}

		sprite := s.Sprites.Get(e)
		if sprite != nil {
			s.renderList = append(s.renderList, renderEntry{
				Entity:    e,
				TextureId: int(sprite.Texture.ID),
				ZIndex:    renderOrder.CalculatedZ,
			})
			return true
		}

		panic("Unknown renderable type")
	})

	slices.SortStableFunc(s.renderList, func(a, b renderEntry) int {
		if a.TextureId == b.TextureId {
			return int(math.Floor(float64(a.ZIndex - b.ZIndex)))
		}
		return int(a.TextureId - b.TextureId)
	})

	// Batch and render
	var currentTex = -1
	for i := range s.renderList {
		entry := &s.renderList[i]
		if entry.TextureId != currentTex || len(s.instanceData) >= 8192 {
			if len(s.instanceData) > 0 {
				s.submitBatch(s.instanceData)
				s.instanceData = s.instanceData[:0]
			}
			currentTex = entry.TextureId
		}
		s.instanceData = append(s.instanceData, s.getInstanceData(entry.Entity))
	}
	s.submitBatch(s.instanceData) // Submit last batch
	s.renderList = s.renderList[:0]

	// ==========
	// DEBUG
	// ==========
	if s.debug {
		rl.BeginMode2D(s.camera)
		s.AABBs.EachEntity(func(e ecs.Entity) bool {
			aabb := s.AABBs.Get(e)
			clr := rl.Green
			isSleeping := s.ColliderSleepStateComponentManager.Get(e)
			if isSleeping != nil {
				clr = rl.Blue
			}
			rl.DrawRectangleLines(int32(aabb.Min.X), int32(aabb.Min.Y), int32(aabb.Max.X-aabb.Min.X), int32(aabb.Max.Y-aabb.Min.Y), clr)
			return true
		})
		s.Collisions.EachEntity(func(entity ecs.Entity) bool {
			pos := s.Positions.Get(entity)
			rl.DrawRectangle(int32(pos.XY.X-8), int32(pos.XY.Y-8), 16, 16, rl.Red)
			return true
		})
		rl.EndMode2D()
	}
}

func (s *RenderSystem) submitBatch(data []stdcomponents.RLTexturePro) {
	rl.BeginMode2D(s.camera)
	if s.debug {
		for i := range data {
			rl.DrawTexturePro(*data[i].Texture, data[i].Frame, data[i].Dest, data[i].Origin, data[i].Rotation, data[i].Tint)
			rl.DrawRectangle(int32(data[i].Dest.X-2), int32(data[i].Dest.Y-2), 4, 4, rl.Red)
		}
	} else {
		for i := range data {
			rl.DrawTexturePro(*data[i].Texture, data[i].Frame, data[i].Dest, data[i].Origin, data[i].Rotation, data[i].Tint)
		}
	}
	rl.EndMode2D()
}

func (s *RenderSystem) getInstanceData(e ecs.Entity) stdcomponents.RLTexturePro {
	return *s.RlTexturePros.Get(e)
}

func (s *RenderSystem) prepareRender(dt time.Duration) {
	wg := new(sync.WaitGroup)
	wg.Add(6)
	s.prepareAnimations(wg)
	go s.prepareFlips(wg)
	go s.preparePositions(wg, dt)
	go s.prepareRotations(wg)
	go s.prepareScales(wg)
	go s.prepareTints(wg)
	wg.Wait()
}

func (s *RenderSystem) prepareAnimations(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.RlTexturePros.Get(entity)
		animation := s.AnimationPlayers.Get(entity)
		if animation == nil {
			return true
		}
		frame := &texturePro.Frame
		if animation.Vertical {
			frame.Y += frame.Height * float32(animation.Current)
		} else {
			frame.X += frame.Width * float32(animation.Current)
		}
		return true
	})
}

func (s *RenderSystem) prepareFlips(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.RlTexturePros.Get(entity)
		flipped := s.Flips.Get(entity)
		if flipped == nil {
			return true
		}
		if flipped.X {
			texturePro.Frame.Width *= -1
		}
		if flipped.Y {
			texturePro.Frame.Height *= -1
		}
		return true
	})
}

func (s *RenderSystem) preparePositions(wg *sync.WaitGroup, dt time.Duration) {
	defer wg.Done()
	dts := dt.Seconds()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.RlTexturePros.Get(entity)
		position := s.Positions.Get(entity)
		if position == nil {
			return true
		}
		decay := 40.0 // DECAY IS TICKRATE DEPENDENT
		x := float32(s.expDecay(float64(texturePro.Dest.X), float64(position.XY.X), decay, dts))
		y := float32(s.expDecay(float64(texturePro.Dest.Y), float64(position.XY.Y), decay, dts))
		texturePro.Dest.X = x
		texturePro.Dest.Y = y
		//player := s.Player.Get(entity)
		//if player != nil {
		//	s.camera.Target.X = x
		//	s.camera.Target.Y = y
		//}

		return true
	})
}

func (s *RenderSystem) prepareRotations(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.RlTexturePros.Get(entity)
		rotation := s.Rotations.Get(entity)
		if rotation == nil {
			return true
		}
		texturePro.Rotation = float32(rotation.Degrees())
		return true
	})
}

func (s *RenderSystem) prepareScales(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.RlTexturePros.Get(entity)
		scale := s.Scales.Get(entity)
		if scale == nil {
			return true
		}
		texturePro.Dest.Width *= scale.XY.X
		texturePro.Dest.Height *= scale.XY.Y
		return true
	})
}

func (s *RenderSystem) prepareTints(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		tr := s.RlTexturePros.Get(entity)
		tint := s.Tints.Get(entity)
		if tint == nil {
			return true
		}
		trTint := &tr.Tint
		trTint.A = tint.A
		trTint.R = tint.R
		trTint.G = tint.G
		trTint.B = tint.B
		return true
	})
}

func (s *RenderSystem) expDecay(a, b, decay, dt float64) float64 {
	return b + (a-b)*(math.Exp(-decay*dt))
}
