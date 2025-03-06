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

type ColliderShape = uint8

const (
	InvalidColliderShape ColliderShape = iota
	BoxColliderShape
	CircleColliderShape
)

type ColliderBox struct {
	Width  float32
	Height float32
}

type ColliderBoxComponentManager = ecs.ComponentManager[ColliderBox]

func NewColliderBoxComponentManager() ColliderBoxComponentManager {
	return ecs.NewComponentManager[ColliderBox](ColliderBoxComponentId)
}

type ColliderCircle struct {
	Radius float32
}

type ColliderCircleComponentManager = ecs.ComponentManager[ColliderCircle]

func NewColliderCircleComponentManager() ColliderCircleComponentManager {
	return ecs.NewComponentManager[ColliderCircle](ColliderCircleComponentId)
}

type GenericCollider struct {
	Shape ColliderShape
}

type GenericColliderComponentManager = ecs.ComponentManager[GenericCollider]

func NewGenericColliderComponentManager() GenericColliderComponentManager {
	return ecs.NewComponentManager[GenericCollider](GenericColliderComponentId)
}
