/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- tema881 Donated 100 RUB

Thank you for your support!
*/

package stdcomponents

import (
	"gomp/pkg/ecs"
	"gomp/vectors"
)

func NewCollisionGrid(collisionLayer CollisionLayer, cellSize float32) CollisionGrid {
	g := CollisionGrid{
		Layer:                  collisionLayer,
		CellSize:               cellSize,
		CellMap:                ecs.NewGenMap[SpatialCellIndex, ecs.Entity](1024),
		CreateCellsAccumulator: nil,
	}

	return g
}

type CollisionGrid struct {
	Layer    CollisionLayer // Layer of the grid
	CellSize float32        // Size of a cell

	CreateCellsAccumulator []ecs.GenMap[SpatialCellIndex, struct{}]
	CellMap                ecs.GenMap[SpatialCellIndex, ecs.Entity]
}

type SpatialCellIndex struct {
	X, Y int
}

func (i SpatialCellIndex) ToVec2() vectors.Vec2 {
	return vectors.Vec2{X: float32(i.X), Y: float32(i.Y)}
}

// Query returns the EntityIds of Cells that intersect the AABB
func (g *CollisionGrid) Query(bb AABB, result []ecs.Entity) []ecs.Entity {
	// get spatial index of aabb
	minSpatialCellIndex := g.GetCellIndex(bb.Min)
	maxSpatialCellIndex := g.GetCellIndex(bb.Max)
	// make a list of all spatial indexes that intersect the aabb
	// get cells that intersect the aabb by spatial indexes
	for i := minSpatialCellIndex.X; i <= maxSpatialCellIndex.X; i++ {
		for j := minSpatialCellIndex.Y; j <= maxSpatialCellIndex.Y; j++ {
			spatialIndex := SpatialCellIndex{X: i, Y: j}
			cellEntity, exists := g.CellMap.Get(spatialIndex)
			if !exists {
				continue
			}
			result = append(result, cellEntity)
		}
	}
	return result
}

func (g *CollisionGrid) GetCellIndex(position vectors.Vec2) SpatialCellIndex {
	return SpatialCellIndex{
		X: int(position.X / g.CellSize),
		Y: int(position.Y / g.CellSize),
	}
}

func (g *CollisionGrid) CalculateSpatialHash(bb AABB) SpatialHash {
	return SpatialHash{
		Min: g.GetCellIndex(bb.Min),
		Max: g.GetCellIndex(bb.Max),
	}
}

type CollisionGridComponentManager = ecs.ComponentManager[CollisionGrid]

func NewCollisionGridComponentManager() CollisionGridComponentManager {
	return ecs.NewComponentManager[CollisionGrid](CollisionGridComponentId)
}
