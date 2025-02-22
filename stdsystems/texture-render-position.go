/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"time"
)

func NewTextureRenderPositionSystem() TextureRenderPositionSystem {
	return TextureRenderPositionSystem{}
}

// TextureRenderPositionSystem is a system that sets Position of textureRender
type TextureRenderPositionSystem struct {
	Positions      *stdcomponents.PositionComponentManager
	TextureRenders *stdcomponents.TextureRenderComponentManager
}

func (s *TextureRenderPositionSystem) Init() {}
func (s *TextureRenderPositionSystem) Run(dt time.Duration) {
	s.TextureRenders.AllParallel(func(entity ecs.Entity, tr *stdcomponents.TextureRender) bool {
		if tr == nil {
			return true
		}

		position := s.Positions.Get(entity)
		if position == nil {
			return true
		}

		decay := 100.0
		tr.Dest.X = float32(expDecay(float64(tr.Dest.X), float64(position.X), decay, dt))
		tr.Dest.Y = float32(expDecay(float64(tr.Dest.Y), float64(position.Y), decay, dt))

		return true
	})
}
func (s *TextureRenderPositionSystem) Destroy() {}

func expDecay(a, b, decay float64, dt time.Duration) float64 {
	return b + (a-b)*(math.Exp(-decay*dt.Seconds()))
}
