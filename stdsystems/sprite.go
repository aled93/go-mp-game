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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

func NewSpriteSystem() SpriteSystem {
	return SpriteSystem{}
}

// SpriteSystem is a system that prepares Sprite to be rendered
type SpriteSystem struct {
	Scales        *stdcomponents.ScaleComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	RLTexturePros *stdcomponents.RLTextureProComponentManager
	Renderables   *stdcomponents.RenderableComponentManager
	RenderOrder   *stdcomponents.RenderOrderComponentManager
}

func (s *SpriteSystem) Init() {}
func (s *SpriteSystem) Run() {
	s.Sprites.EachEntityParallel(func(entity ecs.Entity) bool {
		sprite := s.Sprites.Get(entity)
		scale := s.Scales.Get(entity)
		tr := s.RLTexturePros.Get(entity)

		tr.Texture = sprite.Texture
		tr.Frame = sprite.Frame
		tr.Dest = sprite.Dest
		tr.Origin = rl.Vector2{
			X: sprite.Origin.X * scale.XY.X,
			Y: sprite.Origin.Y * scale.XY.Y,
		}
		tr.Rotation = sprite.Rotation
		tr.Tint = sprite.Tint

		return true
	})
}
func (s *SpriteSystem) Destroy() {}
