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

import (
	"gomp/pkg/ecs"
	"gomp/vectors"
)

type ColliderShape uint8

const (
	InvalidColliderShape ColliderShape = iota
	BoxColliderShape
	CircleColliderShape
	PolygonColliderShape
)

type CollisionMask uint64

func (m CollisionMask) HasLayer(layer CollisionLayer) bool {
	return m&(1<<layer) != 0
}

type CollisionLayer = CollisionMask

const (
	CollisionLayerNone CollisionLayer = 0
)

type BoxCollider struct {
	Width  float32
	Height float32
	Offset vectors.Vec2
	Layer  CollisionLayer
	Mask   CollisionMask
}

type BoxColliderComponentManager = ecs.ComponentManager[BoxCollider]

func NewBoxColliderComponentManager() BoxColliderComponentManager {
	return ecs.NewComponentManager[BoxCollider](ColliderBoxComponentId)
}

type CircleCollider struct {
	Radius float32
	Layer  CollisionLayer
	Mask   CollisionMask
	Offset vectors.Vec2
}

type CircleColliderComponentManager = ecs.ComponentManager[CircleCollider]

func NewCircleColliderComponentManager() CircleColliderComponentManager {
	return ecs.NewComponentManager[CircleCollider](ColliderCircleComponentId)
}

type PolygonCollider struct {
	Vertices []vectors.Vec2
	Layer    CollisionLayer
	Mask     CollisionMask
	Offset   vectors.Vec2
}

type PolygonColliderComponentManager = ecs.ComponentManager[PolygonCollider]

func NewPolygonColliderComponentManager() PolygonColliderComponentManager {
	return ecs.NewComponentManager[PolygonCollider](PolygonColliderComponentId)
}

type GenericCollider struct {
	Shape  ColliderShape
	Layer  CollisionLayer
	Mask   CollisionMask
	Offset vectors.Vec2
}

type GenericColliderComponentManager = ecs.ComponentManager[GenericCollider]

func NewGenericColliderComponentManager() GenericColliderComponentManager {
	return ecs.NewComponentManager[GenericCollider](GenericColliderComponentId)
}
