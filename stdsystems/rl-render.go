/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"sync"
	"time"
)

const (
	batchSize = 1 << 13 // Maximum batch size supported by Raylib
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct {
	EntityManager    *ecs.EntityManager
	RlTexturePros    *stdcomponents.RLTextureProComponentManager
	Positions        *stdcomponents.PositionComponentManager
	AnimationPlayers *stdcomponents.AnimationPlayerComponentManager
	AnimationStates  *stdcomponents.AnimationStateComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	Tints            *stdcomponents.TintComponentManager
	Flips            *stdcomponents.FlipComponentManager
	Renderables      *stdcomponents.RenderableComponentManager
	SpriteMatrixes   *stdcomponents.SpriteMatrixComponentManager
	camera           rl.Camera2D
}

func (s *RenderSystem) Init() {
	rl.InitWindow(1024, 768, "raylib [core] ebiten-ecs - basic window")
	//InitWindow(1024, 768, "raylib [core] ebiten-ecs - basic window")

	s.camera = rl.Camera2D{
		Target:   rl.NewVector2(0, 0),
		Offset:   rl.NewVector2(0, 0),
		Rotation: 0,
		Zoom:     1,
	}
}

func (s *RenderSystem) Run(dt time.Duration) bool {
	if rl.WindowShouldClose() {
		return false
	}

	s.prepareRender(dt)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	rl.BeginMode2D(s.camera)
	s.renderWorld()
	rl.EndMode2D()

	rl.DrawRectangle(0, 0, 200, 60, rl.DarkBrown)
	rl.DrawFPS(10, 10)
	rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)

	rl.EndDrawing()

	return true
}

func (s *RenderSystem) Destroy() {
	rl.CloseWindow()
}

func (s *RenderSystem) renderWorld() {
	s.Renderables.EachEntity(func(entity ecs.Entity) bool {
		renderable := s.Renderables.Get(entity)

		switch *renderable {
		case stdcomponents.SpriteMatrixRenderableType:
			s.renderSpriteMatrix(entity)
		default:
			panic("unknown renderable type")
		}

		return true
	})
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
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
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
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
		texturePro := s.RlTexturePros.Get(entity)
		mirrored := s.Flips.Get(entity)
		if mirrored == nil {
			return true
		}
		if mirrored.X {
			texturePro.Frame.Width *= -1
		}
		if mirrored.Y {
			texturePro.Frame.Height *= -1
		}
		return true
	})
}

func (s *RenderSystem) preparePositions(wg *sync.WaitGroup, dt time.Duration) {
	defer wg.Done()
	//dts := dt.Seconds()
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
		texturePro := s.RlTexturePros.Get(entity)
		position := s.Positions.Get(entity)
		if position == nil {
			return true
		}
		//decay := 16.0 // DECAY IS TICKRATE DEPENDENT
		//texturePro.Dest.X = float32(s.expDecay(float64(texturePro.Dest.X), float64(position.X), decay, dts))
		//texturePro.Dest.Y = float32(s.expDecay(float64(texturePro.Dest.Y), float64(position.Y), decay, dts))
		texturePro.Dest.X = position.X
		texturePro.Dest.Y = position.Y

		return true
	})
}

func (s *RenderSystem) prepareRotations(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
		texturePro := s.RlTexturePros.Get(entity)
		rotation := s.Rotations.Get(entity)
		if rotation == nil {
			return true
		}
		texturePro.Rotation = rotation.Angle
		return true
	})
}

func (s *RenderSystem) prepareScales(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
		texturePro := s.RlTexturePros.Get(entity)
		scale := s.Scales.Get(entity)
		if scale == nil {
			return true
		}
		texturePro.Dest.Width *= scale.X
		texturePro.Dest.Height *= scale.Y
		return true
	})
}

func (s *RenderSystem) prepareTints(wg *sync.WaitGroup) {
	defer wg.Done()
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
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

func (s *RenderSystem) renderSpriteMatrix(entity ecs.Entity) {
	texturePro := s.RlTexturePros.Get(entity)
	rl.DrawTexturePro(*texturePro.Texture, texturePro.Frame, texturePro.Dest, texturePro.Origin, texturePro.Rotation, texturePro.Tint)
}

func (s *RenderSystem) expDecay(a, b, decay, dt float64) float64 {
	return b + (a-b)*(math.Exp(-decay*dt))
}
