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
	Layer       CollisionLayer              // Layer of the grid
	Entities    ecs.PagedArray[ecs.Entity]  // List of Entities in the grid
	ChunkLookup map[SpatialIndex]ecs.Entity // Pointer to cell

	ChunkSize float32
	MinBounds vectors.Vec2
}

func (t *CollisionGrid) RegisterEntity(entity ecs.Entity, aabb *AABB) {
	t.Entities.Append(entity)

	l := aabb.Max.Sub(aabb.Min)

	if l.LengthSquared() < t.MinBounds.LengthSquared() {
		t.MinBounds = l
	}
}

func (t *CollisionGrid) GetSpatialIndex(position vectors.Vec2) SpatialIndex {
	return SpatialIndex{
		X: int(position.X / t.ChunkSize),
		Y: int(position.Y / t.ChunkSize),
	}
}

type CollisionGridComponentManager = ecs.ComponentManager[CollisionGrid]

func NewCollisionGridComponentManager() CollisionGridComponentManager {
	return ecs.NewComponentManager[CollisionGrid](CollisionGridComponentId)
}
