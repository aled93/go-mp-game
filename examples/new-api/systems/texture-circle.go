/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/examples/new-api/components"
	"gomp/pkg/core"
	"gomp/pkg/draw"
	"gomp/pkg/ecs"
	"gomp/pkg/worker"
	"gomp/stdcomponents"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
)

func NewTextureCircleSystem() TextureCircleSystem {
	return TextureCircleSystem{}
}

type TextureCircleSystem struct {
	Circles  *components.PrimitiveCircleComponentManager
	Textures *stdcomponents.RLTextureProComponentManager
	texture  rl.RenderTexture2D

	Engine *core.Engine
}

func (s *TextureCircleSystem) Init() {
	const circleRadius = 128
	var texture = draw.CreateRenderTexture(circleRadius*2, circleRadius*2)
	draw.BeginTextureMode(texture)
	draw.CircleFill(circleRadius, circleRadius, circleRadius, rl.White)
	draw.EndTextureMode()
	s.texture = texture
}

func (s *TextureCircleSystem) Run(dt time.Duration) {
	s.Circles.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		circle := s.Circles.GetUnsafe(entity)
		assert.NotNil(circle, "circle is nil; entity: %d", entity)
		texture := s.Textures.GetUnsafe(entity)
		assert.NotNil(texture, "texture is nil; entity: %d", entity)

		texture.Texture = &s.texture.Texture
		texture.Dest.X = circle.CenterX
		texture.Dest.Y = circle.CenterY
		texture.Dest.Width = circle.Radius * 2
		texture.Dest.Height = circle.Radius * 2
		texture.Frame = rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(s.texture.Texture.Width),
			Height: float32(s.texture.Texture.Height),
		}
		texture.Rotation = circle.Rotation
		texture.Origin.X = circle.Origin.X + circle.Radius
		texture.Origin.Y = circle.Origin.Y + circle.Radius
		texture.Tint = circle.Color
	})
}

func (s *TextureCircleSystem) Destroy() {
	draw.DestroyRenderTexture(s.texture)
}
