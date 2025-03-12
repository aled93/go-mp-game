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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	Renderables   *stdcomponents.RenderableComponentManager
	RenderOrder   *stdcomponents.RenderOrderComponentManager
}

func (s *SpriteSystem) Init() {}
func (s *SpriteSystem) Run() {
	s.Sprites.EachEntityParallel(func(entity ecs.Entity) bool {
		sprite := s.Sprites.Get(entity) //
		position := s.Positions.Get(entity)
		scale := s.Scales.Get(entity)

		if scale == nil {
			scale = &stdcomponents.Scale{
				X: 1,
				Y: 1,
			}
		}

		renderable := s.Renderables.Get(entity)
		if renderable == nil {
			s.Renderables.Create(entity, stdcomponents.SpriteRenderableType)
		}

		renderOrder := s.RenderOrder.Get(entity)
		if renderOrder == nil {
			s.RenderOrder.Create(entity, stdcomponents.RenderOrder{})
		}

		tr := s.RLTexturePros.Get(entity)
		if tr == nil {
			s.RLTexturePros.Create(entity, stdcomponents.RLTexturePro{
				Texture: sprite.Texture, //
				Frame:   sprite.Frame,   //
				Origin: rl.Vector2{
					X: sprite.Origin.X * scale.X,
					Y: sprite.Origin.Y * scale.Y,
				},
				Dest: rl.Rectangle{X: position.X, Y: position.Y, Width: sprite.Frame.Width, Height: sprite.Frame.Height}, //
				Tint: stdcomponents.Tint{
					R: 255,
					G: 255,
					B: 255,
					A: 255,
				},
			})
		} else {
			tr.Dest.Width = sprite.Frame.Width
			tr.Dest.Height = sprite.Frame.Height
		}
		return true
	})
}
func (s *SpriteSystem) Destroy() {}
