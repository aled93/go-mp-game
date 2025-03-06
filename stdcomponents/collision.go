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

type CollisionState uint8

const (
	CollisionStateNone CollisionState = iota
	CollisionStateEnter
	CollisionStateStay
	CollisionStateExit
)

// Collision Marks a proxy entity as representing a collision pair between E1 and E2
type Collision struct {
	E1, E2 ecs.Entity
	State  CollisionState
}

type CollisionComponentManager = ecs.ComponentManager[Collision]

func NewCollisionComponentManager() CollisionComponentManager {
	return ecs.NewComponentManager[Collision](CollisionComponentId)
}
