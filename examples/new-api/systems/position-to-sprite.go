/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"time"
)

// PositionToSpriteSystem updates the sprite's destination based on the position component.
// This is temporary workaround system
type PositionToSpriteSystem struct {
	EntityManager *ecs.EntityManager
	Position      *stdcomponents.PositionComponentManager
	Sprite        *stdcomponents.SpriteComponentManager
	SpriteMatrix  *stdcomponents.SpriteMatrixComponentManager
	Rotation      *stdcomponents.RotationComponentManager
}

func NewPositionToSpriteSystem() PositionToSpriteSystem {
	return PositionToSpriteSystem{}
}

func (s *PositionToSpriteSystem) Init() {
}

func (s *PositionToSpriteSystem) Run(dt time.Duration) {
	s.Sprite.EachEntityParallel(math.MaxInt, func(entity ecs.Entity, _ int) bool {
		position := s.Position.Get(entity)
		sprite := s.Sprite.Get(entity)
		rotation := s.Rotation.Get(entity)

		sprite.Dest.X = position.XY.X
		sprite.Dest.Y = position.XY.Y
		sprite.Rotation = float32(rotation.Degrees())

		return true
	})
	s.SpriteMatrix.EachEntityParallel(math.MaxInt, func(entity ecs.Entity, _ int) bool {
		position := s.Position.Get(entity)
		sprite := s.SpriteMatrix.Get(entity)
		rotation := s.Rotation.Get(entity)

		sprite.Dest.X = position.XY.X
		sprite.Dest.Y = position.XY.Y
		sprite.Rotation = rotation.Angle

		return true
	})
}

func (s *PositionToSpriteSystem) Destroy() {
}
