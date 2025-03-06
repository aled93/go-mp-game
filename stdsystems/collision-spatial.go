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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"runtime"
	"sync"
	"time"
)

func NewSpatialCollisionSystem() SpatialCollisionSystem {
	return SpatialCollisionSystem{
		activeCollisions: make(map[CollisionPair]ecs.Entity),
	}
}

type SpatialCollisionSystem struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	GenericCollider *stdcomponents.GenericColliderComponentManager
	BoxColliders    *stdcomponents.ColliderBoxComponentManager
	Collisions      *stdcomponents.CollisionComponentManager

	activeCollisions    map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions   map[CollisionPair]struct{}
	currentCollisionsMx sync.RWMutex
	activeCollisionsMx  sync.Mutex

	spatialGrid ecs.SpatialGrid
}

func (s *SpatialCollisionSystem) Init() {}
func (s *SpatialCollisionSystem) Run(dt time.Duration) {
	s.currentCollisions = make(map[CollisionPair]struct{})
	s.spatialGrid = ecs.NewSpatialGrid(256)

	s.Positions.EachEntity(func(entity ecs.Entity) bool {
		position := s.Positions.Get(entity)
		s.spatialGrid.AddEntity(entity, float64(position.X), float64(position.Y))
		return true
	})

	// 1. Get all non-empty cells from spatial grid
	cells := s.spatialGrid.GetActiveCells()

	// 2. Process cells in parallel batches
	wg := new(sync.WaitGroup)
	wg.Add(len(cells))

	cellChan := make(chan *ecs.GridCell, len(cells))
	for i := range cells {
		cellChan <- &cells[i]
	}
	defer close(cellChan)

	// Worker pool pattern
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(wg *sync.WaitGroup) {
			for cell := range cellChan {
				s.processCellCollisions(cell)
				wg.Done()
			}
		}(wg)
	}

	wg.Wait()

	s.processExitStates()
}
func (s *SpatialCollisionSystem) Destroy() {}

func (s *SpatialCollisionSystem) processCellCollisions(cell *ecs.GridCell) {
	// Get current cell's entities and nearby cells' entities
	nearbyEntities := s.spatialGrid.GetNearbyEntities(cell.Key)

	// Split into batches within cell
	const entityBatchSize = 50
	for start := 0; start < len(cell.Entities); start += entityBatchSize {
		end := start + entityBatchSize
		if end > len(cell.Entities) {
			end = len(cell.Entities)
		}

		// Process entity batch
		for _, entity := range cell.Entities[start:end] {
			s.checkEntityCollisions(entity, nearbyEntities)
		}
	}
}

func (s *SpatialCollisionSystem) checkEntityCollisions(entity ecs.Entity, candidates []ecs.Entity) {
	// Get components once
	posA := s.Positions.Get(entity)
	colliderA := s.GenericCollider.Get(entity)

	// Check against candidate entities
	for _, other := range candidates {
		//// Ensure each pair is checked only once
		if entity >= other {
			continue
		}

		// Fast ID-based rejection
		if entity == other {
			continue
		}

		//Get other components
		posB := s.Positions.Get(other)

		//Broadphase check
		if !s.spatialGrid.BoundingBoxesIntersect(float64(posA.X), float64(posA.Y), float64(posB.X), float64(posB.Y)) {
			continue
		}

		// Narrowphase check
		switch colliderA.Shape {
		case stdcomponents.BoxColliderShape:
			s.boxToXCollision(entity, other)
		case stdcomponents.CircleColliderShape:
			s.circleToXCollision(entity, other)
		default:
			panic("Unknown collision shape")
		}
	}
}

func (s *SpatialCollisionSystem) registerCollision(entityA, entityB ecs.Entity) {
	pair := CollisionPair{entityA, entityB}.Normalize()

	s.currentCollisionsMx.Lock()
	s.currentCollisions[pair] = struct{}{}
	s.currentCollisionsMx.Unlock()

	s.activeCollisionsMx.Lock()
	defer s.activeCollisionsMx.Unlock()

	// Create proxy entity for new collisions
	if _, exists := s.activeCollisions[pair]; !exists {
		proxy := s.EntityManager.Create()
		s.Collisions.Create(proxy, stdcomponents.Collision{E1: pair.E1, E2: pair.E2, State: stdcomponents.CollisionStateEnter})
		s.activeCollisions[pair] = proxy
	} else {
		s.Collisions.Get(s.activeCollisions[pair]).State = stdcomponents.CollisionStateStay
	}
}

func (s *SpatialCollisionSystem) processExitStates() {
	s.activeCollisionsMx.Lock()
	defer s.activeCollisionsMx.Unlock()
	s.currentCollisionsMx.RLock()
	defer s.currentCollisionsMx.RUnlock()

	for pair, proxy := range s.activeCollisions {
		if _, exists := s.currentCollisions[pair]; !exists {
			collision := s.Collisions.Get(proxy)
			if collision.State == stdcomponents.CollisionStateExit {
				delete(s.activeCollisions, pair)
				s.EntityManager.Delete(proxy)
			} else {
				collision.State = stdcomponents.CollisionStateExit
			}
		}
	}
}

func (s *SpatialCollisionSystem) boxToXCollision(entityA, entityB ecs.Entity) {
	position1 := s.Positions.Get(entityA)
	collider1 := s.BoxColliders.Get(entityA)

	position2 := s.Positions.Get(entityB)
	genericCollider := s.GenericCollider.Get(entityB)
	boxCollider := s.BoxColliders.Get(entityB)

	switch genericCollider.Shape {
	case stdcomponents.BoxColliderShape:
		// Check AABB collision
		if !(position1.X+collider1.Width < position2.X ||
			position1.X > position2.X+boxCollider.Width ||
			position1.Y+collider1.Height < position2.Y ||
			position1.Y > position2.Y+boxCollider.Height) {

			s.registerCollision(entityA, entityB)
		}
	case stdcomponents.CircleColliderShape:
		panic("Circle-Box collision not implemented")
	default:
		panic("Unknown collision shape")
	}
}

func (s *SpatialCollisionSystem) circleToXCollision(entityA, entityB ecs.Entity) {
	panic("Circle-X collision not implemented")
}
