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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"time"
)

func NewCollisionDetectionSystem() CollisionDetectionSystem {
	return CollisionDetectionSystem{}
}

type CollisionDetectionSystem struct {
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

	gridLookup map[stdcomponents.CollisionLayer]ecs.Entity
	numWorkers int
}

func (s *CollisionDetectionSystem) Init() {
	s.gridLookup = make(map[stdcomponents.CollisionLayer]ecs.Entity)
	s.numWorkers = runtime.NumCPU() - 2
}
func (s *CollisionDetectionSystem) Run(dt time.Duration) {
	s.setup()

}
func (s *CollisionDetectionSystem) Destroy() {}
func (s *CollisionDetectionSystem) setup() {
	// Reset grids
	s.CollisionGridComponentManager.EachEntityParallel(s.numWorkers, func(entity ecs.Entity, workerId int) bool {
		grid := s.CollisionGridComponentManager.GetUnsafe(entity)
		assert.NotNil(grid)
		grid.Entities.Reset()
		grid.MinBounds = vectors.Vec2{
			X: math.MaxFloat32,
			Y: math.MaxFloat32,
		}
		return true
	})

	// Accumulate used CollisionLayers
	var collisionLayerAccumulators = make([]stdcomponents.CollisionLayer, s.numWorkers)
	s.GenericCollider.EachEntityParallel(s.numWorkers, func(entity ecs.Entity, workerId int) bool {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)
		collisionLayerAccumulators[workerId] |= 1 << collider.Layer
		return true
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
		gridEntity, exists := s.gridLookup[stdcomponents.CollisionLayer(i)]
		if !exists {
			gridEntity = s.EntityManager.Create()
			s.gridLookup[stdcomponents.CollisionLayer(i)] = gridEntity
			s.CollisionGridComponentManager.Create(gridEntity, stdcomponents.CollisionGrid{
				Layer:       stdcomponents.CollisionLayer(i),
				Entities:    ecs.NewPagedArray[ecs.Entity](),
				ChunkLookup: make(map[stdcomponents.SpatialIndex]ecs.Entity),
				ChunkSize:   0,
				MinBounds: vectors.Vec2{
					X: math.MaxFloat32,
					Y: math.MaxFloat32,
				},
			})
		}
	}

	// Register all entities in grids
	s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
		collider := s.GenericCollider.GetUnsafe(entity)
		assert.NotNil(collider)

		gridEntity := s.gridLookup[collider.Layer]

		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		position := s.Positions.GetUnsafe(entity)
		assert.NotNil(position)

		aabb := s.AABB.GetUnsafe(entity)
		assert.NotNil(aabb)

		grid.RegisterEntity(entity, aabb)

		return true
	})

	// Distribute entities in grids
	s.CollisionGridComponentManager.EachEntity(func(gridEntity ecs.Entity) bool {
		grid := s.CollisionGridComponentManager.GetUnsafe(gridEntity)
		assert.NotNil(grid)

		const chunkScaleFactor = 32
		newChunkSize := grid.MinBounds.Length() * chunkScaleFactor
		newChunkSize = float32(min(int(1)<<(ecs.FastIntLog2(int(newChunkSize))+1), 65535))
		if newChunkSize != grid.ChunkSize {
			for i, c := range grid.ChunkLookup {
				s.EntityManager.Delete(c)
				delete(grid.ChunkLookup, i)
			}

			grid.ChunkSize = newChunkSize
		}

		grid.Entities.EachDataValue(func(entity ecs.Entity) bool {
			position := s.Positions.GetUnsafe(entity)
			assert.NotNil(position)

			spatialIndex := grid.GetSpatialIndex(position.XY)
			chunkEntity, exists := grid.ChunkLookup[spatialIndex]
			if !exists {
				chunkEntity = s.EntityManager.Create()
				grid.ChunkLookup[spatialIndex] = chunkEntity
				s.CollisionChunkComponentManager.Create(chunkEntity, stdcomponents.CollisionChunk{
					Size:  grid.ChunkSize,
					Layer: grid.Layer,
				})
				chunkPos := s.Positions.Create(chunkEntity, stdcomponents.Position{
					XY: spatialIndex.ToVec2().Scale(grid.ChunkSize),
				})
				assert.NotNil(chunkPos)
				s.BvhTreeComponentManager.Create(chunkEntity, stdcomponents.BvhTree{
					Nodes:      ecs.NewSlice[stdcomponents.BvhNode](64),
					AabbNodes:  ecs.NewSlice[stdcomponents.AABB](64),
					Leaves:     ecs.NewSlice[stdcomponents.BvhLeaf](64),
					AabbLeaves: ecs.NewSlice[stdcomponents.AABB](64),
					Codes:      ecs.NewSlice[uint64](64),
					Components: ecs.NewSlice[stdcomponents.BvhComponent](64),
				})
				const colorbase int = 120
				s.Tints.Create(chunkEntity, color.RGBA{
					R: uint8(colorbase + rand.Intn(255-colorbase)),
					G: uint8(colorbase + rand.Intn(255-colorbase)),
					B: uint8(colorbase + rand.Intn(255-colorbase)),
					A: 70,
				})
			}

			tree := s.BvhTreeComponentManager.GetUnsafe(chunkEntity)
			assert.NotNil(tree)

			aabb := s.AABB.GetUnsafe(entity)
			assert.NotNil(aabb)

			tree.AddComponent(entity, *aabb)

			return true
		})
		return true
	})

	s.CollisionChunkComponentManager.EachEntityParallel(s.numWorkers, func(chunkEntity ecs.Entity, workerId int) bool {
		tree := s.BvhTreeComponentManager.GetUnsafe(chunkEntity)
		assert.NotNil(tree)
		tree.Build()
		return true
	})
}
