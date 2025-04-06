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

type CollisionGrid struct {
	Layer      CollisionLayer              // Layer of the grid
	Entities   ecs.PagedArray[ecs.Entity]  // List of Entities in the grid
	CellLookup map[SpatialIndex]ecs.Entity // Pointer to cell

	MinBounds AABB // Bounds of the smallest entity in the grid
	CellSize  vectors.Vec2
}

func (t *CollisionGrid) RegisterEntity(entity ecs.Entity, aabb *AABB) {
	t.Entities.Append(entity)

	if aabb.Min.X < t.MinBounds.Min.X {
		t.MinBounds.Min.X = aabb.Min.X
	}
	if aabb.Min.Y < t.MinBounds.Min.Y {
		t.MinBounds.Min.Y = aabb.Min.Y
	}
}

type CollisionGridComponentManager = ecs.ComponentManager[CollisionGrid]

func NewCollisionGridComponentManager() CollisionGridComponentManager {
	return ecs.NewComponentManager[CollisionGrid](CollisionGridComponentId)
}
