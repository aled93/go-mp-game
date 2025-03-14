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

package stdcomponents

import "gomp/pkg/ecs"

type ColliderShape uint8

const (
	InvalidColliderShape ColliderShape = iota
	BoxColliderShape
	CircleColliderShape
)

type CollisionMask uint64

type CollisionLayer = CollisionMask

const (
	ColliderLayerNone CollisionLayer = 0
)

type BoxCollider struct {
	Width   float32
	Height  float32
	OffsetX float32
	OffsetY float32
	Layer   CollisionLayer
	Mask    CollisionMask
}

type BoxColliderComponentManager = ecs.ComponentManager[BoxCollider]

func NewBoxColliderComponentManager() BoxColliderComponentManager {
	return ecs.NewComponentManager[BoxCollider](ColliderBoxComponentId)
}

type CircleCollider struct {
	Radius  float32
	Layer   CollisionLayer
	Mask    CollisionMask
	OffsetX float32
	OffsetY float32
}

type CircleColliderComponentManager = ecs.ComponentManager[CircleCollider]

func NewCircleColliderComponentManager() CircleColliderComponentManager {
	return ecs.NewComponentManager[CircleCollider](ColliderCircleComponentId)
}

type GenericCollider struct {
	Shape   ColliderShape
	Layer   CollisionLayer
	Mask    CollisionMask
	OffsetX float32
	OffsetY float32
}

type GenericColliderComponentManager = ecs.ComponentManager[GenericCollider]

func NewGenericColliderComponentManager() GenericColliderComponentManager {
	return ecs.NewComponentManager[GenericCollider](GenericColliderComponentId)
}
