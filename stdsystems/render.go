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

package stdsystems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"sort"
	"sync"
	"time"
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct {
	EntityManager    *ecs.EntityManager
	RlTexturePros    *stdcomponents.RLTextureProComponentManager
	Positions        *stdcomponents.PositionComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	AnimationPlayers *stdcomponents.AnimationPlayerComponentManager
	Tints            *stdcomponents.TintComponentManager
	Flips            *stdcomponents.FlipComponentManager
	Renderables      *stdcomponents.RenderableComponentManager
	AnimationStates  *stdcomponents.AnimationStateComponentManager
	SpriteMatrixes   *stdcomponents.SpriteMatrixComponentManager
	renderList       []RenderEntry
	instanceData     []stdcomponents.RLTexturePro
	camera           rl.Camera2D
}

type RenderEntry struct {
	Entity    ecs.Entity
	TextureId int
	ZIndex    float32
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
	s.render()
	rl.DrawRectangle(0, 0, 200, 60, rl.DarkBrown)
	rl.DrawFPS(10, 10)
	rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)
	rl.EndDrawing()

	return true
}

func (s *RenderSystem) Destroy() {
	rl.CloseWindow()
}

func (s *RenderSystem) render() {
	// Extract and sort entities
	if cap(s.renderList) < s.Renderables.Len() {
		s.renderList = append(s.renderList, make([]RenderEntry, 0, s.Renderables.Len()-cap(s.renderList))...)
	}
	s.Renderables.EachEntity(func(e ecs.Entity) bool {
		sprite := s.SpriteMatrixes.Get(e)
		pos := s.Positions.Get(e)
		s.renderList = append(s.renderList, RenderEntry{
			Entity:    e,
			TextureId: int(sprite.Texture.ID),
			ZIndex:    pos.Z,
		})
		return true
	})

	// Sort by texture, then Z-index
	sort.SliceStable(s.renderList, func(i, j int) bool {
		a := s.renderList[i]
		b := s.renderList[j]
		if a.TextureId == b.TextureId {
			return a.ZIndex < b.ZIndex
		}
		return a.TextureId < b.TextureId
	})

	// Batch and render
	var currentTex = -1
	var instanceData []stdcomponents.RLTexturePro = make([]stdcomponents.RLTexturePro, 0, 8192)
	for i := range s.renderList {
		entry := &s.renderList[i]
		if entry.TextureId != currentTex || len(instanceData) >= 8192 {
			if len(instanceData) > 0 {
				s.submitBatch(currentTex, instanceData)
				instanceData = instanceData[:0]
			}
			currentTex = entry.TextureId
		}
		instanceData = append(instanceData, s.getInstanceData(entry.Entity))
	}
	s.submitBatch(currentTex, instanceData) // Submit last batch
	s.renderList = s.renderList[:0]
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

func (s *RenderSystem) submitBatch(texID int, data []stdcomponents.RLTexturePro) {
	rl.BeginMode2D(s.camera)
	for i := range data {
		rl.DrawTexturePro(*data[i].Texture, data[i].Frame, data[i].Dest, data[i].Origin, data[i].Rotation, data[i].Tint)
	}
	rl.EndMode2D()
}
