/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- Монтажер сука Donated 50 RUB

Thank you for your support!
*/

package stdsystems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"runtime"
)

func NewSpriteSystem() SpriteSystem {
	return SpriteSystem{}
}

// SpriteSystem is a system that prepares Sprite to be rendered
type SpriteSystem struct {
	Positions     *stdcomponents.PositionComponentManager
	Scales        *stdcomponents.ScaleComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	RLTexturePros *stdcomponents.RLTextureProComponentManager
	RenderOrder   *stdcomponents.RenderOrderComponentManager

	numWorkers int
}

func (s *SpriteSystem) Init() {
	s.numWorkers = runtime.NumCPU() - 2
}
func (s *SpriteSystem) Run() {
	var accRenderOrder = make([][]ecs.Entity, s.numWorkers)
	var accRLTexturePros = make([][]ecs.Entity, s.numWorkers)
	s.Sprites.EachEntityParallel(s.numWorkers)(func(entity ecs.Entity, workerId int) bool {
		renderOrder := s.RenderOrder.GetUnsafe(entity)
		if renderOrder == nil {
			accRenderOrder[workerId] = append(accRenderOrder[workerId], entity)
		}
		tr := s.RLTexturePros.GetUnsafe(entity)
		if tr == nil {
			accRLTexturePros[workerId] = append(accRLTexturePros[workerId], entity)
		}
		return true
	})
	for a := range accRenderOrder {
		for _, entity := range accRenderOrder[a] {
			s.RenderOrder.Create(entity, stdcomponents.RenderOrder{})
		}
	}
	for a := range accRLTexturePros {
		for _, entity := range accRLTexturePros[a] {
			s.RLTexturePros.Create(entity, stdcomponents.RLTexturePro{})
		}
	}

	s.Sprites.EachEntityParallel(s.numWorkers)(func(entity ecs.Entity, _ int) bool {
		sprite := s.Sprites.GetUnsafe(entity)
		assert.NotNil(sprite)

		position := s.Positions.GetUnsafe(entity)
		assert.NotNil(position)

		scale := s.Scales.GetUnsafe(entity)
		assert.NotNil(scale)

		renderOrder := s.RenderOrder.GetUnsafe(entity)
		assert.NotNil(renderOrder)

		tr := s.RLTexturePros.GetUnsafe(entity)
		assert.NotNil(tr)

		tr.Texture = sprite.Texture
		tr.Frame = sprite.Frame
		tr.Origin = rl.Vector2{
			X: sprite.Origin.X * scale.XY.X,
			Y: sprite.Origin.Y * scale.XY.Y,
		}
		tr.Dest.X = position.XY.X
		tr.Dest.Y = position.XY.Y
		tr.Dest.Width = sprite.Frame.Width
		tr.Dest.Height = sprite.Frame.Height
		tr.Tint = sprite.Tint
		return true
	})
}
func (s *SpriteSystem) Destroy() {}
