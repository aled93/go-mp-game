/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/core"
	"gomp/pkg/ecs"
	"gomp/pkg/kbd"
	"gomp/pkg/util"
	"gomp/pkg/worker"
	"gomp/stdcomponents"
	"math"
	"runtime"
	"time"

	"github.com/negrel/assert"
)

func NewTexturePositionSmoothSystem() TexturePositionSmoothSystem {
	return TexturePositionSmoothSystem{}
}

type TexturePositionSmoothSystem struct {
	TexturePositionSmooth *stdcomponents.TexturePositionSmoothComponentManager
	Position              *stdcomponents.PositionComponentManager
	RLTexture             *stdcomponents.RLTextureProComponentManager
	numWorkers            int
	Engine                *core.Engine
}

func (s *TexturePositionSmoothSystem) Init() {
	s.numWorkers = runtime.NumCPU() - 2
}

func (s *TexturePositionSmoothSystem) Run(dt time.Duration) {
	//DEBUG Temporary, TODO: remove
	if kbd.IsKeyPressed(kbd.KeycodeI) {
		s.TexturePositionSmooth.ProcessComponents(func(t *stdcomponents.TexturePositionSmooth, workerId worker.WorkerId) {
			*t = stdcomponents.TexturePositionSmoothOff
		})
	}
	if kbd.IsKeyPressed(kbd.KeycodeO) {
		s.TexturePositionSmooth.ProcessComponents(func(t *stdcomponents.TexturePositionSmooth, workerId worker.WorkerId) {
			*t = stdcomponents.TexturePositionSmoothLerp
		})
	}
	if kbd.IsKeyPressed(kbd.KeycodeP) {
		s.TexturePositionSmooth.ProcessComponents(func(t *stdcomponents.TexturePositionSmooth, workerId worker.WorkerId) {
			*t = stdcomponents.TexturePositionSmoothExpDecay
		})
	}
	//END DEBUG

	s.TexturePositionSmooth.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		position := s.Position.GetUnsafe(entity)
		texture := s.RLTexture.GetUnsafe(entity)
		smooth := s.TexturePositionSmooth.GetUnsafe(entity)
		if texture == nil {
			return
		}
		assert.Nil(position, "position is nil")

		switch *smooth {
		case stdcomponents.TexturePositionSmoothLerp:
			dest := util.NewVec2(texture.Dest.X, texture.Dest.Y)
			xy := dest.Lerp(position.XY, util.Scalar(dt))
			texture.Dest.X = xy.X
			texture.Dest.Y = xy.Y
		case stdcomponents.TexturePositionSmoothExpDecay:
			dest := util.NewVec2(texture.Dest.X, texture.Dest.Y)
			xy := s.ExpDecay2D(dest, position.XY, 10, float64(dt))
			texture.Dest.X = xy.X
			texture.Dest.Y = xy.Y
		default:
			panic("not implemented")
		}
	})
}

func (s *TexturePositionSmoothSystem) Destroy() {}

// ExpDecay2D applies an exponential decay to the position vector.
// TODO: float32 math package
func (_ *TexturePositionSmoothSystem) ExpDecay2D(a, b util.Vec2, decay, dt float64) util.Vec2 {
	return util.NewVec2(
		b.X+(a.X-b.X)*(float32(math.Exp(-decay*dt))),
		b.Y+(a.Y-b.Y)*(float32(math.Exp(-decay*dt))),
	)
}
