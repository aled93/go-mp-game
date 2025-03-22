/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- rpecb Donated 500 RUB

Thank you for your support!
*/

package stdsystems

import "C"
import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"time"
)

func NewColliderSystem() ColliderSystem {
	return ColliderSystem{}
}

type ColliderSystem struct {
	EntityManager    *ecs.EntityManager
	Positions        *stdcomponents.PositionComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	GenericColliders *stdcomponents.GenericColliderComponentManager
	BoxColliders     *stdcomponents.BoxColliderComponentManager
	CircleColliders  *stdcomponents.CircleColliderComponentManager
	AABB             *stdcomponents.AABBComponentManager
}

func (s *ColliderSystem) Init() {}
func (s *ColliderSystem) Run(dt time.Duration) {
	s.BoxColliders.EachEntity(func(entity ecs.Entity) bool {
		boxCollider := s.BoxColliders.Get(entity)

		genCollider := s.GenericColliders.Get(entity)
		if genCollider == nil {
			genCollider = s.GenericColliders.Create(entity, stdcomponents.GenericCollider{})
		}
		genCollider.Layer = boxCollider.Layer
		genCollider.Mask = boxCollider.Mask
		genCollider.Offset.X = boxCollider.Offset.X
		genCollider.Offset.Y = boxCollider.Offset.Y
		genCollider.Shape = stdcomponents.BoxColliderShape

		position := s.Positions.Get(entity)
		scale := s.Scales.Get(entity)
		rotation := s.Rotations.Get(entity)
		aabb := s.AABB.Get(entity)
		if aabb == nil {
			aabb = s.AABB.Create(entity, stdcomponents.AABB{})
		}

		a := boxCollider.WH
		b := vectors.Vec2{X: 0, Y: boxCollider.WH.Y}
		c := vectors.Vec2{X: 0, Y: 0}
		d := vectors.Vec2{X: boxCollider.WH.X, Y: 0}

		c = c.Sub(boxCollider.Offset).Rotate(rotation.Angle)
		a = a.Sub(boxCollider.Offset).Rotate(rotation.Angle)
		b = b.Sub(boxCollider.Offset).Rotate(rotation.Angle)
		d = d.Sub(boxCollider.Offset).Rotate(rotation.Angle)

		aabb.Min = vectors.Vec2{X: min(b.X, c.X, a.X, d.X), Y: min(b.Y, c.Y, a.Y, d.Y)}.Mul(scale.XY)
		aabb.Max = vectors.Vec2{X: max(b.X, c.X, a.X, d.X), Y: max(b.Y, c.Y, a.Y, d.Y)}.Mul(scale.XY)

		aabb.Min = position.XY.Add(aabb.Min)
		aabb.Max = position.XY.Add(aabb.Max)

		return true
	})

	s.CircleColliders.EachEntity(func(entity ecs.Entity) bool {
		circleCollider := s.CircleColliders.Get(entity)

		genCollider := s.GenericColliders.Get(entity)
		if genCollider == nil {
			s.GenericColliders.Create(entity, stdcomponents.GenericCollider{
				Shape: stdcomponents.CircleColliderShape,
				Layer: circleCollider.Layer,
				Mask:  circleCollider.Mask,
				Offset: vectors.Vec2{
					X: circleCollider.Offset.X,
					Y: circleCollider.Offset.Y,
				},
			})
		}

		position := s.Positions.Get(entity)
		scale := s.Scales.Get(entity)
		aabb := s.AABB.Get(entity)
		if aabb == nil {
			aabb = s.AABB.Create(entity, stdcomponents.AABB{})
		}

		aabb.Min = position.XY.Sub(circleCollider.Offset.Mul(scale.XY))
		aabb.Max = position.XY.Add(circleCollider.Offset.Scale(-1).SubScalar(circleCollider.Radius).Mul(scale.XY))

		return true
	})
}
func (s *ColliderSystem) Destroy() {}
