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
	"time"
)

func NewCollisionDetectionSystem() CollisionDetectionSystem {
	return CollisionDetectionSystem{
		cellSizeX: 96,
		cellSizeY: 128,
	}
}

type CollisionDetectionSystem struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	GenericCollider *stdcomponents.GenericColliderComponentManager
	BoxColliders    *stdcomponents.ColliderBoxComponentManager
	Collisions      *stdcomponents.CollisionComponentManager
	SpatialIndex    *stdcomponents.SpatialIndexComponentManager

	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
	cellSizeX         int
	cellSizeY         int

	spatialBuckets map[stdcomponents.SpatialIndex][]ecs.Entity
	aabbs          map[ecs.Entity]AABB
	entityToCell   map[ecs.Entity]stdcomponents.SpatialIndex
}

type AABB struct {
	Left, Right, Top, Bottom float32
}

func (s *CollisionDetectionSystem) Init() {
	s.activeCollisions = make(map[CollisionPair]ecs.Entity)
}
func (s *CollisionDetectionSystem) Run(dt time.Duration) {
	s.entityToCell = make(map[ecs.Entity]stdcomponents.SpatialIndex, s.GenericCollider.Len())
	s.spatialBuckets = make(map[stdcomponents.SpatialIndex][]ecs.Entity, 32)

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

	s.currentCollisions = make(map[CollisionPair]struct{})

	s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
		collider := s.GenericCollider.Get(entity)

		switch collider.Shape {
		case stdcomponents.BoxColliderShape:
			s.boxToXCollision(entity)
		case stdcomponents.CircleColliderShape:
			s.circleToXCollision(entity)
		default:
			panic("Unknown collider shape")
		}
		return true
	})

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

func (s *CollisionDetectionSystem) boxToXCollision(entityA ecs.Entity) {
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
			if genericCollider1.Mask&(1<<genericCollider2.Layer) == 0 && genericCollider2.Mask&(1<<genericCollider1.Layer) == 0 {
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

				s.registerCollision(entityA, entityB, (position1.X+position2.X)/2, (position1.Y+position2.Y)/2)
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
