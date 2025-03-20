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
		aabb := s.AABB.Get(entity)
		if aabb == nil {
			aabb = s.AABB.Create(entity, stdcomponents.AABB{})
		}
		aabb.Min = vectors.Vec2{
			X: position.X - (boxCollider.Offset.X * scale.X),
			Y: position.Y - (boxCollider.Offset.Y * scale.Y),
		}
		aabb.Max = vectors.Vec2{
			X: position.X + (boxCollider.Width-boxCollider.Offset.X)*scale.X,
			Y: position.Y + (boxCollider.Height-boxCollider.Offset.Y)*scale.Y,
		}

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

		aabb.Min = vectors.Vec2{
			X: position.X - (circleCollider.Offset.X * scale.X),
			Y: position.Y + (circleCollider.Radius-circleCollider.Offset.Y)*scale.Y,
		}
		aabb.Max = vectors.Vec2{
			X: position.X + (circleCollider.Radius-circleCollider.Offset.X)*scale.X,
			Y: position.Y - (circleCollider.Offset.Y * scale.Y),
		}

		return true
	})
}
func (s *ColliderSystem) Destroy() {}
