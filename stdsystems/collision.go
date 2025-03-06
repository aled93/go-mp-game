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
	"sync"
	"time"
)

func NewCollisionSystem() CollisionSystem {
	return CollisionSystem{}
}

type CollisionSystem struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	GenericCollider *stdcomponents.GenericColliderComponentManager
	BoxColliders    *stdcomponents.ColliderBoxComponentManager
	Collisions      *stdcomponents.CollisionComponentManager

	activeCollisions    map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions   map[CollisionPair]struct{}
	currentCollisionsMx sync.RWMutex
	activeCollisionsMx  sync.Mutex
}

type CollisionPair struct {
	E1, E2 ecs.Entity
}

func (c CollisionPair) Normalize() CollisionPair {
	if c.E1 > c.E2 {
		return CollisionPair{c.E2, c.E1}
	}
	return c
}

func (s *CollisionSystem) Init() {
	s.activeCollisions = make(map[CollisionPair]ecs.Entity)
}
func (s *CollisionSystem) Run(dt time.Duration) {
	s.currentCollisions = make(map[CollisionPair]struct{})

	entities := s.GenericCollider.RawEntities(make([]ecs.Entity, 0, s.GenericCollider.Len()))

	const batchSize = 64 // Tune this based on performance testing
	wg := new(sync.WaitGroup)

	// Process entities in batches
	for start := 0; start < len(entities); start += batchSize {
		end := start + batchSize
		if end > len(entities) {
			end = len(entities)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				entity := entities[i]
				collision := s.GenericCollider.Get(entity)

				switch collision.Shape {
				case stdcomponents.BoxColliderShape:
					s.boxToXCollision(entities, i)
				case stdcomponents.CircleColliderShape:
					s.circleToXCollision(entities, i)
				default:
					panic("Unknown collision shape")
				}
			}
		}(start, end)
	}

	wg.Wait()

	s.processExitStates()
}
func (s *CollisionSystem) Destroy() {}

func (s *CollisionSystem) registerCollision(entityA, entityB ecs.Entity) {
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

func (s *CollisionSystem) processExitStates() {
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

func (s *CollisionSystem) boxToXCollision(entities []ecs.Entity, i int) {
	entityA := entities[i]
	position1 := s.Positions.Get(entityA)
	collider1 := s.BoxColliders.Get(entityA)

	for j := i + 1; j < len(entities); j++ {
		entityB := entities[j]
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
}

func (s *CollisionSystem) circleToXCollision(entities []ecs.Entity, i int) {
	panic("Circle-X collision not implemented")
}
