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
	"github.com/negrel/assert"
	"gomp/pkg/core"
	"gomp/pkg/ecs"
	"gomp/pkg/worker"
	"gomp/stdcomponents"
	"gomp/vectors"
	"image/color"
	"math"
	"math/rand"
	"time"
)

const (
	collidersPerCell = 64
)

func NewCollisionSetupSystem() CollisionSetupSystem {
	return CollisionSetupSystem{}
}

type CollisionSetupSystem struct {
	EntityManager                      *ecs.EntityManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	Tints                              *stdcomponents.TintComponentManager
	GenericCollider                    *stdcomponents.GenericColliderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	PolygonColliders                   *stdcomponents.PolygonColliderComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	SpatialIndex                       *stdcomponents.SpatialIndexComponentManager
	AABB                               *stdcomponents.AABBComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTreeComponentManager            *stdcomponents.BvhTreeComponentManager
	CollisionGridComponentManager      *stdcomponents.CollisionGridComponentManager
	CollisionChunkComponentManager     *stdcomponents.CollisionChunkComponentManager
	CollisionCellComponentManager      *stdcomponents.CollisionCellComponentManager

	gridLookup map[stdcomponents.CollisionLayer]ecs.Entity
	Engine     *core.Engine
}

func (s *CollisionSetupSystem) Init() {
	s.gridLookup = make(map[stdcomponents.CollisionLayer]ecs.Entity)
}
func (s *CollisionSetupSystem) Run(dt time.Duration) {
	// Reset grids
	s.CollisionGridComponentManager.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		grid := s.CollisionGridComponentManager.GetUnsafe(entity)
		assert.NotNil(grid)
		grid.Entities.Reset()
		grid.MinBounds = vectors.Vec2{
			X: math.MaxFloat32,
			Y: math.MaxFloat32,
		}
	})

	// Accumulate used CollisionLayers
	var collisionLayerAccumulators = make([]stdcomponents.CollisionLayer, s.Engine.Pool().NumWorkers())
	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)
		collisionLayerAccumulators[workerId] |= 1 << collider.Layer
	})
	collisionLayerAccumulator := stdcomponents.CollisionLayer(0)
	for _, mask := range collisionLayerAccumulators {
		collisionLayerAccumulator |= mask
	}

	// For each bit in collisionLayerAccumulators, create a grid if it doesn't exist
	for i := uint(0); i < 32; i++ {
		if collisionLayerAccumulator&(1<<i) == 0 {
			continue
		}
		collisionLayer := stdcomponents.CollisionLayer(i)
		gridEntity, exists := s.gridLookup[collisionLayer]
		if exists {
			continue
		}

		gridEntity = s.EntityManager.Create()
		s.gridLookup[collisionLayer] = gridEntity
		grid := s.CollisionGridComponentManager.Create(gridEntity, stdcomponents.CollisionGrid{
			Layer:       collisionLayer,
			Entities:    ecs.NewPagedArray[ecs.Entity](),
			ChunkLookup: make(map[stdcomponents.SpatialIndex]ecs.Entity),
			ChunkSize:   0,
			MinBounds: vectors.Vec2{
				X: math.MaxFloat32,
				Y: math.MaxFloat32,
			},
		})
		grid.Init(collisionLayer, s.Engine.Pool())
	}

	// Calculate cell size for each grid
	s.BoxColliders.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		boxCollider := s.BoxColliders.GetUnsafe(entity)
		assert.NotNil(boxCollider)

		gridEntity := s.gridLookup[boxCollider.Layer]
		assert.NotNil(gridEntity)

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		grid.CellSizeAccumulator[workerId] = min(grid.CellSizeAccumulator[workerId], boxCollider.WH.Length())
	})
	s.CircleColliders.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		circleCollider := s.CircleColliders.GetUnsafe(entity)
		assert.NotNil(circleCollider)

		gridEntity := s.gridLookup[circleCollider.Layer]
		assert.NotNil(gridEntity)

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		grid.CellSizeAccumulator[workerId] = min(grid.CellSizeAccumulator[workerId], circleCollider.Radius*2)
	})
	s.CollisionGridComponentManager.EachEntity()(func(entity ecs.Entity) bool {
		grid := s.CollisionGridComponentManager.GetUnsafe(entity)
		assert.NotNil(grid)

		for i := range grid.CellSizeAccumulator {
			grid.CellSize = min(grid.CellSize, grid.CellSizeAccumulator[i])
		}
		grid.CellSize *= collidersPerCell
		return true
	})

	// Create spatial hash components
	var spatialHashAccumulator = make([][]ecs.Entity, s.Engine.Pool().NumWorkers())
	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		if s.SpatialIndex.Has(entity) {
			return
		}

		spatialHashAccumulator[id] = append(spatialHashAccumulator[id], entity)
	})
	for i := range spatialHashAccumulator {
		acc := spatialHashAccumulator[i]
		for j := range acc {
			entity := acc[j]
			s.SpatialIndex.Create(entity, stdcomponents.SpatialIndex{})
		}
	}

	// Calculate spatial index for each entity
	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)

		gridEntity, ok := s.gridLookup[collider.Layer]
		assert.True(ok)

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		position := s.Positions.GetUnsafe(entity)
		assert.NotNil(position)

		spatialIndex := s.SpatialIndex.GetUnsafe(entity)
		assert.NotNil(spatialIndex)

		newIndex := grid.GetSpatialIndex(position.XY)
		spatialIndex.X = newIndex.X
		spatialIndex.Y = newIndex.Y
	})

	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)

		gridEntity, ok := s.gridLookup[collider.Layer]
		assert.True(ok)

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		spatialIndex := s.SpatialIndex.GetUnsafe(entity)
		assert.NotNil(spatialIndex)

		_, exists := grid.CellLookup[*spatialIndex]
		if exists {
			return
		}

		grid.CellAccumulator[id][*spatialIndex] = struct{}{}
	})

	// Create cells
	s.CollisionGridComponentManager.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		grid := s.CollisionGridComponentManager.GetUnsafe(entity)
		assert.NotNil(grid)

		for i := range grid.CellAccumulator {
			for spatialIndex := range grid.CellAccumulator[i] {
				cellEntity := s.EntityManager.Create()
				grid.Cells.Append(cellEntity)
				grid.CellLookup[spatialIndex] = grid.Cells.Len() - 1
				collisionCell := s.CollisionCellComponentManager.Create(cellEntity, stdcomponents.CollisionCell{})
				collisionCell.Init(grid.CellSize, grid.Layer, s.Engine.Pool())
				position := s.Positions.Create(cellEntity, stdcomponents.Position{
					XY: spatialIndex.ToVec2().Scale(grid.CellSize),
				})
				assert.NotNil(position)
				s.BvhTreeComponentManager.Create(cellEntity, stdcomponents.BvhTree{
					Nodes:      ecs.NewPagedArray[stdcomponents.BvhNode](),
					AabbNodes:  ecs.NewPagedArray[stdcomponents.AABB](),
					Leaves:     ecs.NewPagedArray[stdcomponents.BvhLeaf](),
					AabbLeaves: ecs.NewPagedArray[stdcomponents.AABB](),
					Codes:      ecs.NewPagedArray[uint64](),
					Components: ecs.NewPagedArray[stdcomponents.BvhComponent](),
				})
				const colorbase int = 120
				s.Tints.Create(cellEntity, color.RGBA{
					R: uint8(colorbase + rand.Intn(255-colorbase)),
					G: uint8(colorbase + rand.Intn(255-colorbase)),
					B: uint8(colorbase + rand.Intn(255-colorbase)),
					A: 70,
				})
			}
			grid.CellAccumulator[i] = make(map[stdcomponents.SpatialIndex]struct{})
		}
	})

	// Distribute entities in grid cells
	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)

		gridEntity, ok := s.gridLookup[collider.Layer]
		assert.True(ok)

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		spatialIndex := s.SpatialIndex.GetUnsafe(entity)
		assert.NotNil(spatialIndex)

		cellEntityLookup, exists := grid.CellLookup[*spatialIndex]
		assert.True(exists)
		cellEntity := grid.Cells.GetValue(cellEntityLookup)

		cell := s.CollisionCellComponentManager.GetUnsafe(cellEntity)
		assert.NotNil(cell)

		_, exists = cell.MemberLookup.Get(entity)
		if exists {
			return
		}
		cell.InputAccumulator[id].Append(entity)
	})

	//Build grid cells
	//s.CollisionCellComponentManager.ProcessComponents(func(cell *stdcomponents.CollisionCell, id worker.WorkerId) {
	//	for i := range cell.InputAccumulator {
	//		for e := range cell.InputAccumulator[i].EachDataValue() {
	//			cell.AddMember(e)
	//		}
	//		cell.InputAccumulator[i].Reset()
	//	}
	//})

	//Build bvh trees
	s.CollisionCellComponentManager.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		cell := s.CollisionCellComponentManager.GetUnsafe(entity)
		assert.NotNil(cell)

		bvhTree := s.BvhTreeComponentManager.GetUnsafe(entity)
		assert.NotNil(bvhTree)

		for i := range cell.InputAccumulator {
			for e := range cell.InputAccumulator[i].EachDataValue() {
				aabb := s.AABB.GetUnsafe(e)
				assert.NotNil(aabb)
				bvhTree.AddComponent(e, *aabb)
			}
			cell.InputAccumulator[i].Reset()
		}
	})
	s.CollisionCellComponentManager.ProcessEntities(func(entity ecs.Entity, id worker.WorkerId) {
		bvhTree := s.BvhTreeComponentManager.GetUnsafe(entity)
		assert.NotNil(bvhTree)

		bvhTree.Build()
	})
}
func (s *CollisionSetupSystem) Destroy() {
	s.gridLookup = nil
}
