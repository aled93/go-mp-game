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
	"sync"
	"time"
)

const (
	chunkScaleFactor = 32
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
}

func (s *CollisionDetectionSystem) Init() {
	s.gridLookup = make(map[stdcomponents.CollisionLayer]ecs.Entity)
}
func (s *CollisionDetectionSystem) Run(dt time.Duration) {
	s.setup()

}
func (s *CollisionDetectionSystem) Destroy() {}
func (s *CollisionDetectionSystem) setup() {
	// Reset grids
	s.CollisionGridComponentManager.EachEntity(func(entity ecs.Entity) bool {
		grid := s.CollisionGridComponentManager.Get(entity)
		assert.NotNil(grid)

		grid.Entities.Reset()
		grid.MinBounds = vectors.Vec2{
			X: math.MaxFloat32,
			Y: math.MaxFloat32,
		}

		return true
	})

	// Register all entities in grids
	s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
		collider := s.GenericCollider.Get(entity)
		assert.NotNil(collider)

		gridEntity, exists := s.gridLookup[collider.Layer]
		if !exists {
			gridEntity = s.EntityManager.Create()
			s.gridLookup[collider.Layer] = gridEntity
			s.CollisionGridComponentManager.Create(gridEntity, stdcomponents.CollisionGrid{
				Layer:       collider.Layer,
				Entities:    ecs.NewPagedArray[ecs.Entity](),
				ChunkLookup: make(map[stdcomponents.SpatialIndex]ecs.Entity),
				ChunkSize:   0,
				MinBounds: vectors.Vec2{
					X: math.MaxFloat32,
					Y: math.MaxFloat32,
				},
			})
		}

		grid := s.CollisionGridComponentManager.Get(gridEntity)
		assert.NotNil(grid)

		position := s.Positions.Get(entity)
		assert.NotNil(position)

		aabb := s.AABB.Get(entity)
		assert.NotNil(aabb)

		grid.RegisterEntity(entity, aabb)

		return true
	})

	// Distribute entities in grids
	s.CollisionGridComponentManager.EachEntity(func(gridEntity ecs.Entity) bool {
		grid := s.CollisionGridComponentManager.Get(gridEntity)
		assert.NotNil(grid)

		newChunkSize := grid.MinBounds.Length() * chunkScaleFactor
		newChunkSize = float32(min(int(1)<<(ecs.FastIntLog2(int(newChunkSize))+1), 65535))
		if newChunkSize != grid.ChunkSize {
			for i, c := range grid.ChunkLookup {
				s.EntityManager.Delete(c)
				delete(grid.ChunkLookup, i)
			}

			grid.ChunkSize = newChunkSize
		}

		grid.Entities.AllDataValue(func(entity ecs.Entity) bool {
			position := s.Positions.Get(entity)
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
					Nodes:           ecs.NewPagedArray[stdcomponents.BvhNode](),
					AabbNodes:       ecs.NewPagedArray[stdcomponents.AABB](),
					Leaves:          ecs.NewPagedArray[stdcomponents.BvhLeaf](),
					AabbLeaves:      ecs.NewPagedArray[stdcomponents.AABB](),
					Codes:           ecs.NewPagedArray[uint64](),
					Components:      ecs.NewPagedArray[stdcomponents.BvhComponent](),
					ComponentsSlice: make([]stdcomponents.BvhComponent, 0, grid.Entities.Len()),
				})
				s.Tints.Create(chunkEntity, color.RGBA{
					R: uint8(rand.Intn(255)),
					G: uint8(rand.Intn(255)),
					B: uint8(rand.Intn(255)),
					A: 20,
				})
			}

			tree := s.BvhTreeComponentManager.Get(chunkEntity)
			assert.NotNil(tree)

			aabb := s.AABB.Get(entity)
			assert.NotNil(aabb)

			tree.AddComponent(entity, *aabb)

			return true
		})
		return true
	})

	// Build BVH trees
	wg := &sync.WaitGroup{}
	s.CollisionChunkComponentManager.EachEntity(func(gridEntity ecs.Entity) bool {
		tree := s.BvhTreeComponentManager.Get(gridEntity)
		assert.NotNil(tree)

		wg.Add(1)
		go func() {
			defer wg.Done()
			tree.Build()
		}()

		return true
	})
	wg.Wait()
}
