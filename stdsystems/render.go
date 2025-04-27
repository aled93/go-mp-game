/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- HromRu Donated 1 500 RUB
<- MuTaToR Donated 500 RUB

Thank you for your support!
*/

package stdsystems

import (
	"gomp/pkg/draw"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"slices"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct {
	EntityManager *ecs.EntityManager
	FrameBuffer2D *stdcomponents.FrameBuffer2DComponentManager

	renderTextures []rl.RenderTexture2D
	frames         []stdcomponents.FrameBuffer2D

	monitorWidth  int
	monitorHeight int
}

func (s *RenderSystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
	s.renderTextures = make([]rl.RenderTexture2D, 0, s.FrameBuffer2D.Len())
	s.frames = make([]stdcomponents.FrameBuffer2D, 0, s.FrameBuffer2D.Len())
}

func (s *RenderSystem) Run(dt time.Duration) bool {
	s.FrameBuffer2D.EachComponent()(func(c *stdcomponents.FrameBuffer2D) bool {
		s.frames = append(s.frames, *c)
		return true
	})
	slices.SortFunc(s.frames, func(a, b stdcomponents.FrameBuffer2D) int {
		return int(a.Layer - b.Layer)
	})

	for _, frame := range s.frames {
		draw.BeginBlendMode(frame.BlendMode)
		draw.Texture(frame.Texture.Texture,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(frame.Texture.Texture.Width),
				Height: -float32(frame.Texture.Texture.Height),
			},
			frame.Dst,
			rl.Vector2{},
			frame.Rotation,
			frame.Tint,
		)
		draw.EndBlendMode()
	}

	s.frames = s.frames[:0]
	return false
}

func (s *RenderSystem) Destroy() {
}

type RenderInjector struct {
	EntityManager *ecs.EntityManager
	FrameBuffer2D *stdcomponents.FrameBuffer2DComponentManager
}

func (s *RenderSystem) InjectWorld(injector *RenderInjector) {
	s.EntityManager = injector.EntityManager
	s.FrameBuffer2D = injector.FrameBuffer2D
}
