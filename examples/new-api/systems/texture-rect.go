/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/examples/new-api/components"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"time"
)

func NewTextureRectSystem() TextureRectSystem {
	return TextureRectSystem{}
}

type TextureRectSystem struct {
	TextureRect *components.TextureRectComponentManager
	Textures    *stdcomponents.RLTextureProComponentManager
	texture     rl.RenderTexture2D
}

func (s *TextureRectSystem) Init() {
	var texture = rl.LoadRenderTexture(1, 1) // :)
	rl.BeginTextureMode(texture)
	rl.ClearBackground(rl.White)
	rl.EndTextureMode()
	s.texture = texture
}

func (s *TextureRectSystem) Run(dt time.Duration) {
	// Create shallow copy of texture to draw rectangles
	s.TextureRect.EachEntityParallel(math.MaxInt, func(entity ecs.Entity, i int) bool {
		rect := s.TextureRect.Get(entity)
		texture := s.Textures.Get(entity)
		assert.Nil(texture, "texture is nil; entity: %d", entity)

		texture.Texture = &s.texture.Texture
		texture.Dest = rect.Dest
		texture.Rotation = rect.Rotation
		texture.Origin = rect.Origin
		texture.Tint = rect.Color
		return true
	})
}

func (s *TextureRectSystem) Destroy() {
	rl.UnloadRenderTexture(s.texture)
}
