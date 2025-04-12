/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"cmp"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"slices"
	"sync"
	"time"
)

func NewRender2DCamerasSystem() Render2DCamerasSystem {
	return Render2DCamerasSystem{}
}

type Render2DCamerasSystem struct {
	Renderables      *stdcomponents.RenderableComponentManager
	RenderVisibles   *stdcomponents.RenderVisibleComponentManager
	RenderOrders     *stdcomponents.RenderOrderComponentManager
	Textures         *stdcomponents.RLTextureProComponentManager
	Tints            *stdcomponents.TintComponentManager
	AABBs            *stdcomponents.AABBComponentManager
	Cameras          *stdcomponents.CameraComponentManager
	RenderTexture2D  *stdcomponents.FrameBuffer2DComponentManager
	AnimationPlayers *stdcomponents.AnimationPlayerComponentManager
	Flips            *stdcomponents.FlipComponentManager
	Positions        *stdcomponents.PositionComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	renderObjects    []renderObject
}

type renderObject struct {
	texture stdcomponents.RLTexturePro
	mask    stdcomponents.CameraLayer
	order   float32
}

func (s *Render2DCamerasSystem) Init() {
	s.renderObjects = make([]renderObject, 0, s.RenderVisibles.Len())
}

func (s *Render2DCamerasSystem) Run(dt time.Duration) {
	s.prepareRender(dt)

	s.Cameras.EachEntity(func(entity ecs.Entity) bool {
		camera := s.Cameras.Get(entity)
		renderTexture := s.RenderTexture2D.Get(entity)

		// Collect and sort render objects
		s.RenderVisibles.EachEntity(func(entity ecs.Entity) bool {
			r := s.Renderables.Get(entity)
			t := s.Textures.Get(entity)
			o := s.RenderOrders.Get(entity)

			//TODO: rework this with future new assets manager
			if t != nil && t.Texture != nil {
				s.renderObjects = append(s.renderObjects, renderObject{
					texture: *t,
					mask:    r.CameraMask,
					order:   o.CalculatedZ,
				})
			}

			return true
		})

		slices.SortFunc(s.renderObjects, func(a, b renderObject) int {
			return cmp.Compare(a.order, b.order)
		})

		// Draw render objects
		rl.BeginTextureMode(renderTexture.Texture)
		rl.BeginMode2D(camera.Camera2D)
		rl.ClearBackground(camera.BGColor)

		for _, obj := range s.renderObjects {
			if camera.Layer&obj.mask != 0 {
				assert.Nil(obj.texture, "EntityTexturePro is nil")
				rl.DrawTexturePro(*obj.texture.Texture, obj.texture.Frame, obj.texture.Dest, obj.texture.Origin, obj.texture.Rotation, obj.texture.Tint)
			}
		}

		rl.EndMode2D()
		rl.EndTextureMode()

		s.renderObjects = s.renderObjects[:0]
		return true
	})
}

func (s *Render2DCamerasSystem) Destroy() {
}

func (s *Render2DCamerasSystem) prepareRender(dt time.Duration) {
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

func (s *Render2DCamerasSystem) prepareAnimations(wg *sync.WaitGroup) {
	defer wg.Done()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.Textures.Get(entity)
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

func (s *Render2DCamerasSystem) prepareFlips(wg *sync.WaitGroup) {
	defer wg.Done()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.Textures.Get(entity)
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

func (s *Render2DCamerasSystem) preparePositions(wg *sync.WaitGroup, dt time.Duration) {
	defer wg.Done()
	//dts := dt.Seconds()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.Textures.Get(entity)
		position := s.Positions.Get(entity)
		if position == nil {
			return true
		}
		//decay := 40.0 // DECAY IS TICKRATE DEPENDENT
		//x := float32(s.expDecay(float64(texturePro.Dest.X), float64(position.XY.X), decay, dts))
		//y := float32(s.expDecay(float64(texturePro.Dest.Y), float64(position.XY.Y), decay, dts))
		texturePro.Dest.X = position.XY.X
		texturePro.Dest.Y = position.XY.Y

		return true
	})
}

func (s *Render2DCamerasSystem) prepareRotations(wg *sync.WaitGroup) {
	defer wg.Done()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.Textures.Get(entity)
		rotation := s.Rotations.Get(entity)
		if rotation == nil {
			return true
		}
		texturePro.Rotation = float32(rotation.Degrees())
		return true
	})
}

func (s *Render2DCamerasSystem) prepareScales(wg *sync.WaitGroup) {
	defer wg.Done()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		texturePro := s.Textures.Get(entity)
		scale := s.Scales.Get(entity)
		if scale == nil {
			return true
		}
		texturePro.Dest.Width *= scale.XY.X
		texturePro.Dest.Height *= scale.XY.Y
		return true
	})
}

func (s *Render2DCamerasSystem) prepareTints(wg *sync.WaitGroup) {
	defer wg.Done()
	s.Textures.EachEntityParallel(128, func(entity ecs.Entity, workerId int) bool {
		tr := s.Textures.Get(entity)
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

func (s *Render2DCamerasSystem) expDecay(a, b, decay, dt float64) float64 {
	return b + (a-b)*(math.Exp(-decay*dt))
}
