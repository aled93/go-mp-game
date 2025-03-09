/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)dw

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

func NewCollisionDetectionSystem() CollisionDetectionSystem {
	return CollisionDetectionSystem{
		cellSizeX:      96,
		cellSizeY:      128,
		spatialBuckets: make(map[stdcomponents.SpatialIndex][]ecs.Entity, 32),
	}
}

type CollisionDetectionSystem struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	GenericCollider *stdcomponents.GenericColliderComponentManager
	BoxColliders    *stdcomponents.ColliderBoxComponentManager
	Collisions      *stdcomponents.CollisionComponentManager
	SpatialIndex    *stdcomponents.SpatialIndexComponentManager

	cellSizeX int
	cellSizeY int

	// Cache
	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
	spatialBuckets    map[stdcomponents.SpatialIndex][]ecs.Entity
	entityToCell      map[ecs.Entity]stdcomponents.SpatialIndex
	aabbs             map[ecs.Entity]AABB
}

type AABB struct {
	Left, Right, Top, Bottom float32
}

func (s *CollisionDetectionSystem) Init() {
	s.activeCollisions = make(map[CollisionPair]ecs.Entity)
}

type CollisionEvent struct {
	entityA, entityB ecs.Entity
	posX, posY       float32
}

func (s *CollisionDetectionSystem) Run(dt time.Duration) {
	if len(s.entityToCell) < s.GenericCollider.Len() {
		s.entityToCell = make(map[ecs.Entity]stdcomponents.SpatialIndex, s.GenericCollider.Len())
	}
	// Reuse spatialBuckets to reduce allocations
	for k := range s.spatialBuckets {
		delete(s.spatialBuckets, k)
	}
	s.currentCollisions = make(map[CollisionPair]struct{})

	// Build spatial buckets and entity-to-cell map
	s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
		position := s.Positions.Get(entity)
		cellX := int(position.X) / s.cellSizeX
		cellY := int(position.Y) / s.cellSizeY
		cell := stdcomponents.SpatialIndex{X: cellX, Y: cellY}
		s.entityToCell[entity] = cell
		s.spatialBuckets[cell] = append(s.spatialBuckets[cell], entity)
		return true
	})

	// Precompute AABBs for box colliders
	s.aabbs = make(map[ecs.Entity]AABB, s.BoxColliders.Len())
	s.BoxColliders.EachEntity(func(entity ecs.Entity) bool {
		position := s.Positions.Get(entity)
		collider := s.BoxColliders.Get(entity)
		s.aabbs[entity] = AABB{
			Left:   position.X,
			Right:  position.X + collider.Width,
			Top:    position.Y,
			Bottom: position.Y + collider.Height,
		}
		return true
	})

	// Create collision channel
	collisionChan := make(chan CollisionEvent, 4096)
	doneChan := make(chan struct{})

	// Start result collector
	go func() {
		for event := range collisionChan {
			pair := CollisionPair{event.entityA, event.entityB}.Normalize()
			s.currentCollisions[pair] = struct{}{}

			if _, exists := s.activeCollisions[pair]; !exists {
				proxy := s.EntityManager.Create()
				s.Collisions.Create(proxy, stdcomponents.Collision{E1: pair.E1, E2: pair.E2, State: stdcomponents.CollisionStateEnter})
				s.Positions.Create(proxy, stdcomponents.Position{X: event.posX, Y: event.posY})
				s.activeCollisions[pair] = proxy
			} else {
				proxy := s.activeCollisions[pair]
				s.Collisions.Get(proxy).State = stdcomponents.CollisionStateStay
				s.Positions.Get(proxy).X = event.posX
				s.Positions.Get(proxy).Y = event.posY
			}
		}
		close(doneChan)
	}()

	entities := s.GenericCollider.RawEntities(make([]ecs.Entity, 0, s.GenericCollider.Len()))

	// Worker pool setup
	var wg sync.WaitGroup
	maxNumWorkers := runtime.NumCPU() * 4
	entitiesLength := len(entities)
	// get minimum 1 worker for small amount of entities, and maximum maxNumWorkers for a lot entities
	numWorkers := max(min(entitiesLength/10, maxNumWorkers), 1)
	chunkSize := entitiesLength / numWorkers

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		startIndex := i * chunkSize
		endIndex := startIndex + chunkSize
		if i == numWorkers-1 { // have to set endIndex to entites lenght, if last worker
			endIndex = entitiesLength
		}

		go func(start int, end int) {
			defer wg.Done()

			for _, entityA := range entities[start:end] {
				collider := s.GenericCollider.Get(entityA)

				switch collider.Shape {
				case stdcomponents.BoxColliderShape:
					s.boxToXCollision(entityA, collisionChan)
				case stdcomponents.CircleColliderShape:
					s.circleToXCollision(entityA)
				default:
					panic("Unknown collider shape")
				}
			}
		}(startIndex, endIndex)
	}

	// Wait for workers and close collision channel
	wg.Wait()
	close(collisionChan)
	<-doneChan // Wait for result collector

	//s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
	//	collider := s.GenericCollider.Get(entity)
	//
	//	switch collider.Shape {
	//	case stdcomponents.BoxColliderShape:
	//		s.boxToXCollision(entity)
	//	case stdcomponents.CircleColliderShape:
	//		s.circleToXCollision(entity)
	//	default:
	//		panic("Unknown collider shape")
	//	}
	//	return true
	//})

	s.processExitStates()
}
func (s *CollisionDetectionSystem) Destroy() {}

func (s *CollisionDetectionSystem) registerCollision(entityA, entityB ecs.Entity, posX, posY float32) {
	pair := CollisionPair{entityA, entityB}.Normalize()

	s.currentCollisions[pair] = struct{}{}

	// Create proxy entity for new collisions
	if _, exists := s.activeCollisions[pair]; !exists {
		proxy := s.EntityManager.Create()
		s.Collisions.Create(proxy, stdcomponents.Collision{E1: pair.E1, E2: pair.E2, State: stdcomponents.CollisionStateEnter})
		s.Positions.Create(proxy, stdcomponents.Position{X: posX, Y: posY})
		s.activeCollisions[pair] = proxy
	} else {
		s.Collisions.Get(s.activeCollisions[pair]).State = stdcomponents.CollisionStateStay
		s.Positions.Get(s.activeCollisions[pair]).X = posX
		s.Positions.Get(s.activeCollisions[pair]).Y = posY
	}
}

func (s *CollisionDetectionSystem) processExitStates() {
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

func (s *CollisionDetectionSystem) boxToXCollision(entityA ecs.Entity, collisionChan chan<- CollisionEvent) {
	position1 := s.Positions.Get(entityA)
	spatialIndex1 := s.entityToCell[entityA]
	genericCollider1 := s.GenericCollider.Get(entityA)

	var nearByIndexes = [9]stdcomponents.SpatialIndex{
		{spatialIndex1.X, spatialIndex1.Y},
		{spatialIndex1.X - 1, spatialIndex1.Y},
		{spatialIndex1.X + 1, spatialIndex1.Y},
		{spatialIndex1.X, spatialIndex1.Y - 1},
		{spatialIndex1.X, spatialIndex1.Y + 1},
		{spatialIndex1.X - 1, spatialIndex1.Y - 1},
		{spatialIndex1.X + 1, spatialIndex1.Y + 1},
		{spatialIndex1.X - 1, spatialIndex1.Y + 1},
		{spatialIndex1.X + 1, spatialIndex1.Y - 1},
	}

	for _, spatialIndex := range nearByIndexes {
		bucket := s.spatialBuckets[spatialIndex]
		for _, entityB := range bucket {
			if entityA >= entityB {
				continue // Skip duplicate checks
			}

			// Broad Phase
			genericCollider2 := s.GenericCollider.Get(entityB)
			if genericCollider1.Mask&(1<<genericCollider2.Layer) == 0 &&
				genericCollider2.Mask&(1<<genericCollider1.Layer) == 0 {
				continue
			}

			// Narrow Phase
			position2 := s.Positions.Get(entityB)

			switch genericCollider2.Shape {
			case stdcomponents.BoxColliderShape:
				// Inside boxToXCollision
				aabbA := s.aabbs[entityA]
				aabbB := s.aabbs[entityB]

				// Check X-axis first
				if aabbA.Right < aabbB.Left || aabbA.Left > aabbB.Right {
					continue
				}
				// Then Y-axis
				if aabbA.Bottom < aabbB.Top || aabbA.Top > aabbB.Bottom {
					continue
				}

				posX := (position1.X + position2.X) / 2
				posY := (position1.Y + position2.Y) / 2
				collisionChan <- CollisionEvent{entityA, entityB, posX, posY}
			case stdcomponents.CircleColliderShape:
				panic("Circle-Box collision not implemented")
			default:
				panic("Unknown collision shape")
			}
		}
	}
}

func (s *CollisionDetectionSystem) circleToXCollision(entityA ecs.Entity) {
	panic("Circle-X collision not implemented")
}
