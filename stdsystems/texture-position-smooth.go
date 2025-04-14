/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
	"runtime"
	"time"
)

func NewTexturePositionSmoothSystem() TexturePositionSmoothSystem {
	return TexturePositionSmoothSystem{}
}

type TexturePositionSmoothSystem struct {
	TexturePositionSmooth *stdcomponents.TexturePositionSmoothComponentManager
	Position              *stdcomponents.PositionComponentManager
	RLTexture             *stdcomponents.RLTextureProComponentManager
	numWorkers            int
}

func (s *TexturePositionSmoothSystem) Init() {
	s.numWorkers = runtime.NumCPU() - 2
}

func (s *TexturePositionSmoothSystem) Run(dt time.Duration) {
	//DEBUG Temporary, TODO: remove
	if rl.IsKeyPressed(rl.KeyI) {
		s.TexturePositionSmooth.EachComponentParallel(s.numWorkers)(func(t *stdcomponents.TexturePositionSmooth, i int) bool {
			*t = stdcomponents.TexturePositionSmoothOff
			return true
		})
	}
	if rl.IsKeyPressed(rl.KeyO) {
		s.TexturePositionSmooth.EachComponentParallel(s.numWorkers)(func(t *stdcomponents.TexturePositionSmooth, i int) bool {
			*t = stdcomponents.TexturePositionSmoothLerp
			return true
		})
	}
	if rl.IsKeyPressed(rl.KeyP) {
		s.TexturePositionSmooth.EachComponentParallel(s.numWorkers)(func(t *stdcomponents.TexturePositionSmooth, i int) bool {
			*t = stdcomponents.TexturePositionSmoothExpDecay
			return true
		})
	}
	//END DEBUG

	s.TexturePositionSmooth.EachEntityParallel(s.numWorkers)(func(entity ecs.Entity, i int) bool {
		position := s.Position.GetUnsafe(entity)
		texture := s.RLTexture.GetUnsafe(entity)
		smooth := s.TexturePositionSmooth.GetUnsafe(entity)
		if texture == nil {
			return true
		}
		assert.Nil(position, "position is nil")

		switch *smooth {
		case stdcomponents.TexturePositionSmoothLerp:
			dest := vectors.Vec2{X: texture.Dest.X, Y: texture.Dest.Y}
			xy := s.Lerp2D(dest, position.XY, dt)
			texture.Dest.X = xy.X
			texture.Dest.Y = xy.Y
		case stdcomponents.TexturePositionSmoothExpDecay:
			dest := vectors.Vec2{X: texture.Dest.X, Y: texture.Dest.Y}
			xy := s.ExpDecay2D(dest, position.XY, 10, float64(dt))
			texture.Dest.X = xy.X
			texture.Dest.Y = xy.Y
		default:
		}

		return true
	})
}

func (s *TexturePositionSmoothSystem) Destroy() {}

func (_ *TexturePositionSmoothSystem) Lerp2D(a, b vectors.Vec2, dt time.Duration) vectors.Vec2 {
	return a.Add(b.Sub(a).Scale(float32(dt)))
}

// ExpDecay2D applies an exponential decay to the position vector.
// TODO: float32 math package
func (_ *TexturePositionSmoothSystem) ExpDecay2D(a, b vectors.Vec2, decay, dt float64) vectors.Vec2 {
	return vectors.Vec2{
		X: b.X + (a.X-b.X)*(float32(math.Exp(-decay*dt))),
		Y: b.Y + (a.Y-b.Y)*(float32(math.Exp(-decay*dt))),
	}
}
