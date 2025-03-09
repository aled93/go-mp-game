/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewColliderSystem() ColliderSystem {
	return ColliderSystem{}
}

type ColliderSystem struct {
	EntityManager    *ecs.EntityManager
	GenericColliders *stdcomponents.GenericColliderComponentManager
	BoxColliders     *stdcomponents.BoxColliderComponentManager
	CircleColliders  *stdcomponents.CircleColliderComponentManager
}

func (s *ColliderSystem) Init() {}
func (s *ColliderSystem) Run(dt time.Duration) {
	s.BoxColliders.EachEntity(func(entity ecs.Entity) bool {
		boxCollider := s.BoxColliders.Get(entity)

		genCollider := s.GenericColliders.Get(entity)
		if genCollider == nil {
			s.GenericColliders.Create(entity, stdcomponents.GenericCollider{
				Shape:   stdcomponents.BoxColliderShape,
				Layer:   boxCollider.Layer,
				Mask:    boxCollider.Mask,
				OffsetX: boxCollider.OffsetX,
				OffsetY: boxCollider.OffsetY,
			})
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
			})
		}

		return true
	})
}
func (s *ColliderSystem) Destroy() {}
