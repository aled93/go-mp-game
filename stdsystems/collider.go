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
	EntityManager                      *ecs.EntityManager
	Positions                          *stdcomponents.PositionComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Velocities                         *stdcomponents.VelocityComponentManager
	GenericColliders                   *stdcomponents.GenericColliderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	AABB                               *stdcomponents.AABBComponentManager
}

func (s *ColliderSystem) Init() {}
func (s *ColliderSystem) Run(dt time.Duration) {
	s.BoxColliders.EachEntity(func(entity ecs.Entity) bool {
		boxCollider := s.BoxColliders.GetUnsafe(entity)

		genCollider := s.GenericColliders.GetUnsafe(entity)
		if genCollider == nil {
			genCollider = s.GenericColliders.Create(entity, stdcomponents.GenericCollider{})
		}
		genCollider.Layer = boxCollider.Layer
		genCollider.Mask = boxCollider.Mask
		genCollider.Offset.X = boxCollider.Offset.X
		genCollider.Offset.Y = boxCollider.Offset.Y
		genCollider.Shape = stdcomponents.BoxColliderShape
		genCollider.AllowSleep = boxCollider.AllowSleep

		position := s.Positions.GetUnsafe(entity)
		scale := s.Scales.GetUnsafe(entity)
		rotation := s.Rotations.GetUnsafe(entity)
		aabb := s.AABB.GetUnsafe(entity)
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
		circleCollider := s.CircleColliders.GetUnsafe(entity)

		genCollider := s.GenericColliders.GetUnsafe(entity)
		if genCollider == nil {
			genCollider = s.GenericColliders.Create(entity, stdcomponents.GenericCollider{})
		}

		genCollider.Layer = circleCollider.Layer
		genCollider.Mask = circleCollider.Mask
		genCollider.Offset.X = circleCollider.Offset.X
		genCollider.Offset.Y = circleCollider.Offset.Y
		genCollider.Shape = stdcomponents.CircleColliderShape
		genCollider.AllowSleep = circleCollider.AllowSleep

		position := s.Positions.GetUnsafe(entity)
		scale := s.Scales.GetUnsafe(entity)
		aabb := s.AABB.GetUnsafe(entity)
		if aabb == nil {
			aabb = s.AABB.Create(entity, stdcomponents.AABB{})
		}

		offset := circleCollider.Offset.Mul(scale.XY)
		scaledRadius := scale.XY.Scale(circleCollider.Radius)
		aabb.Min = position.XY.Add(offset).Sub(scaledRadius)
		aabb.Max = position.XY.Add(offset).Add(scaledRadius)

		return true
	})

	s.GenericColliders.EachEntity(func(entity ecs.Entity) bool {
		genCollider := s.GenericColliders.GetUnsafe(entity)

		if genCollider.AllowSleep {
			shouldSleep := true
			velocity := s.Velocities.GetUnsafe(entity)
			if velocity != nil {
				if velocity.Vec2().LengthSquared() != 0 {
					shouldSleep = false
				}
			}
			isSleeping := s.ColliderSleepStateComponentManager.GetUnsafe(entity)
			if shouldSleep {
				if isSleeping == nil {
					isSleeping = s.ColliderSleepStateComponentManager.Create(entity, stdcomponents.ColliderSleepState{})
				}
			} else {
				if isSleeping != nil {
					s.ColliderSleepStateComponentManager.Delete(entity)
				}
			}
		}
		return true
	})
}
func (s *ColliderSystem) Destroy() {}
