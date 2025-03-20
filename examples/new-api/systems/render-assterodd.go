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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"slices"
	"sync"
	"time"
)

func NewRenderAssteroddSystem() RenderAssteroddSystem {
	return RenderAssteroddSystem{
		instanceData: make([]stdcomponents.RLTexturePro, 0, 8192),
	}
}

type RenderAssteroddSystem struct {
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
	Sprites          *stdcomponents.SpriteComponentManager
	SpriteMatrixes   *stdcomponents.SpriteMatrixComponentManager
	RenderOrders     *stdcomponents.RenderOrderComponentManager
	ColliderBoxes    *stdcomponents.BoxColliderComponentManager
	Collisions       *stdcomponents.CollisionComponentManager
	renderList       []renderEntry
	instanceData     []stdcomponents.RLTexturePro
	camera           rl.Camera2D
	SceneManager     *components.AsteroidSceneManagerComponentManager

	monitorWidth  int
	monitorHeight int

	Player *components.PlayerTagComponentManager
}

func (s *RenderAssteroddSystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
	s.camera = rl.Camera2D{
		Target:   rl.NewVector2(float32(s.monitorWidth/2), float32(s.monitorHeight/2)),
		Offset:   rl.NewVector2(float32(s.monitorWidth/2), float32(s.monitorHeight/2)),
		Rotation: 0,
		Zoom:     1,
	}
}
func (s *RenderAssteroddSystem) Run(dt time.Duration) bool {
	if rl.WindowShouldClose() {
		return false
	}

	s.prepareRender(dt)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	s.render()

	rl.DrawFPS(10, 10)
	rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)
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

	rl.EndDrawing()

	return true
}

func (s *RenderAssteroddSystem) Destroy() {}

func (s *RenderAssteroddSystem) render() {

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
				s.submitBatch(currentTex, s.instanceData)
				s.instanceData = s.instanceData[:0]
			}
			currentTex = entry.TextureId
		}
		s.instanceData = append(s.instanceData, s.getInstanceData(entry.Entity))
	}
	s.submitBatch(currentTex, s.instanceData) // Submit last batch
	s.renderList = s.renderList[:0]

	rl.BeginMode2D(s.camera)
	s.Collisions.EachEntity(func(entity ecs.Entity) bool {
		pos := s.Positions.Get(entity)
		rl.DrawRectangle(int32(pos.X-8), int32(pos.Y-8), 16, 16, rl.Red)
		return true
	})
	s.ColliderBoxes.EachEntity(func(e ecs.Entity) bool {
		box := s.ColliderBoxes.Get(e)
		pos := s.Positions.Get(e)
		scale := s.Scales.Get(e)
		col := s.ColliderBoxes.Get(e)

		rl.DrawRectangleLines(int32(pos.X-(col.Offset.X*scale.X)), int32(pos.Y-(col.Offset.Y*scale.Y)), int32(box.Width*scale.X), int32(box.Height*scale.Y), rl.DarkGreen)
		return true
	})
	s.Renderables.EachEntity(func(e ecs.Entity) bool {
		position := s.Positions.Get(e)
		rl.DrawRectangle(int32(position.X-2), int32(position.Y-2), 4, 4, rl.Red)
		return true
	})
	rl.EndMode2D()
}

func (s *RenderAssteroddSystem) submitBatch(texID int, data []stdcomponents.RLTexturePro) {
	rl.BeginMode2D(s.camera)
	for i := range data {
		rl.DrawTexturePro(*data[i].Texture, data[i].Frame, data[i].Dest, data[i].Origin, data[i].Rotation, data[i].Tint)
	}
	rl.EndMode2D()
}

func (s *RenderAssteroddSystem) getInstanceData(e ecs.Entity) stdcomponents.RLTexturePro {
	return *s.RlTexturePros.Get(e)
}

func (s *RenderAssteroddSystem) prepareRender(dt time.Duration) {
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

func (s *RenderAssteroddSystem) prepareAnimations(wg *sync.WaitGroup) {
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

func (s *RenderAssteroddSystem) prepareFlips(wg *sync.WaitGroup) {
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

func (s *RenderAssteroddSystem) preparePositions(wg *sync.WaitGroup, dt time.Duration) {
	defer wg.Done()
	dts := dt.Seconds()
	s.RlTexturePros.EachEntityParallel(func(entity ecs.Entity) bool {
		texturePro := s.RlTexturePros.Get(entity)
		position := s.Positions.Get(entity)
		if position == nil {
			return true
		}
		decay := 40.0 // DECAY IS TICKRATE DEPENDENT
		x := float32(s.expDecay(float64(texturePro.Dest.X), float64(position.X), decay, dts))
		y := float32(s.expDecay(float64(texturePro.Dest.Y), float64(position.Y), decay, dts))
		texturePro.Dest.X = x
		texturePro.Dest.Y = y
		player := s.Player.Get(entity)
		if player != nil {
			s.camera.Target.X = x
			s.camera.Target.Y = y
		}

		return true
	})
}

func (s *RenderAssteroddSystem) prepareRotations(wg *sync.WaitGroup) {
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

func (s *RenderAssteroddSystem) prepareScales(wg *sync.WaitGroup) {
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

func (s *RenderAssteroddSystem) prepareTints(wg *sync.WaitGroup) {
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

func (s *RenderAssteroddSystem) expDecay(a, b, decay, dt float64) float64 {
	return b + (a-b)*(math.Exp(-decay*dt))
}
