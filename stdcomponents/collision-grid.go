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
	"gomp/pkg/worker"
	"gomp/vectors"
	"math"
)

type CollisionGrid struct {
	Entities    ecs.PagedArray[ecs.Entity]  // List of Entities in the grid
	ChunkLookup map[SpatialIndex]ecs.Entity // Pointer to cell
	ChunkSize   float32
	MinBounds   vectors.Vec2

	// NEW API
	Layer               CollisionLayer              // Layer of the grid
	Cells               ecs.PagedArray[ecs.Entity]  // List of Cells in the grid
	CellLookup          map[SpatialIndex]int        // Index to cell in Cells
	CellSizeAccumulator []float32                   // Accumulator for cell size
	CellAccumulator     []map[SpatialIndex]struct{} // Accumulator for Cells
	CellSize            float32                     // Size of a cell
}

func (g *CollisionGrid) Init(collisionLayer CollisionLayer, pool *worker.Pool) {
	g.Layer = collisionLayer
	g.CellSize = math.MaxFloat32
	g.Cells = ecs.NewPagedArray[ecs.Entity]()
	g.CellLookup = make(map[SpatialIndex]int)
	g.CellSizeAccumulator = make([]float32, pool.NumWorkers())
	g.CellAccumulator = make([]map[SpatialIndex]struct{}, pool.NumWorkers())
	for i := 0; i < pool.NumWorkers(); i++ {
		g.CellSizeAccumulator[i] = math.MaxFloat32
		g.CellAccumulator[i] = make(map[SpatialIndex]struct{})
	}
}

// Query returns the EntityIds of Cells that intersect the AABB
func (g *CollisionGrid) Query(bb AABB, result []ecs.Entity) []ecs.Entity {
	// get spatial index of aabb
	minSpatialIndex := g.GetSpatialIndex(bb.Min)
	maxSpatialIndex := g.GetSpatialIndex(bb.Max)
	// make a list of all spatial indexes that intersect the aabb
	// get cells that intersect the aabb by spatial indexes
	for i := minSpatialIndex.X; i <= maxSpatialIndex.X; i++ {
		for j := minSpatialIndex.Y; j <= maxSpatialIndex.Y; j++ {
			spatialIndex := SpatialIndex{X: i, Y: j}
			cellIndex, exists := g.CellLookup[spatialIndex]
			if !exists {
				continue

			}
			cell := g.Cells.GetValue(cellIndex)
			result = append(result, cell)
		}
	}
	return result
}

func (g *CollisionGrid) RegisterEntity(entity ecs.Entity, aabb *AABB) {
	g.Entities.Append(entity)

	l := aabb.Max.Sub(aabb.Min)

	if l.LengthSquared() < g.MinBounds.LengthSquared() {
		g.MinBounds = l
	}
}

func (g *CollisionGrid) GetSpatialIndex(position vectors.Vec2) SpatialIndex {
	return SpatialIndex{
		X: int(position.X / g.CellSize),
		Y: int(position.Y / g.CellSize),
	}
}

type CollisionGridComponentManager = ecs.ComponentManager[CollisionGrid]

func NewCollisionGridComponentManager() CollisionGridComponentManager {
	return ecs.NewComponentManager[CollisionGrid](CollisionGridComponentId)
}
