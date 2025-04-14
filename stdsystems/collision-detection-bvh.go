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
	"gomp/pkg/bvh"
	gjk "gomp/pkg/collision"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"runtime"
	"sync"
	"time"
)

const debugTree = false

var maxNumWorkers = runtime.NumCPU() - 1

func NewCollisionDetectionBVHSystem() CollisionDetectionBVHSystem {
	return CollisionDetectionBVHSystem{
		activeCollisions: make(map[CollisionPair]ecs.Entity),
		collisionEvents:  make([]ecs.PagedArray[CollisionEvent], maxNumWorkers),
		trees:            make([]bvh.Tree, 0, 8),
		treesLookup:      make(map[stdcomponents.CollisionLayer]int, 8),
	}
}

type CollisionDetectionBVHSystem struct {
	EntityManager                      *ecs.EntityManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	GenericCollider                    *stdcomponents.GenericColliderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	PolygonColliders                   *stdcomponents.PolygonColliderComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	SpatialIndex                       *stdcomponents.SpatialIndexComponentManager
	AABB                               *stdcomponents.AABBComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTreeComponentManager            *stdcomponents.BvhTreeComponentManager

	trees       []bvh.Tree
	treesLookup map[stdcomponents.CollisionLayer]int

	collisionEvents []ecs.PagedArray[CollisionEvent]

	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
	entities          []ecs.Entity
}

func (s *CollisionDetectionBVHSystem) Init() {
	for i := range maxNumWorkers {
		s.collisionEvents[i] = ecs.NewPagedArray[CollisionEvent]()
	}
}

func (s *CollisionDetectionBVHSystem) Run(dt time.Duration) {
	s.currentCollisions = make(map[CollisionPair]struct{})
	defer s.processExitStates()

	if s.GenericCollider.Len() == 0 {
		return
	}

	// Fill trees
	s.GenericCollider.EachEntity()(func(entity ecs.Entity) bool {
		aabb := s.AABB.GetUnsafe(entity)
		layer := s.GenericCollider.GetUnsafe(entity).Layer

		treeId, exists := s.treesLookup[layer]
		if !exists {
			treeId = len(s.trees)
			s.trees = append(s.trees, bvh.NewTree(layer))
			s.treesLookup[layer] = treeId
		}

		s.trees[treeId].AddComponent(entity, *aabb)

		return true
	})

	// Build trees
	wg := new(sync.WaitGroup)
	wg.Add(len(s.trees))
	for i := range s.trees {
		go func(i int, w *sync.WaitGroup) {
			defer w.Done()
			s.trees[i].Build()
		}(i, wg)
	}
	wg.Wait()

	if len(s.entities) < s.GenericCollider.Len() {
		s.entities = make([]ecs.Entity, 0, s.GenericCollider.Len())
	}
	s.entities = s.GenericCollider.RawEntities(s.entities)
	s.findEntityCollisions(s.entities)
	s.registerCollisionEvents()
}

func (s *CollisionDetectionBVHSystem) Destroy() {}

func (s *CollisionDetectionBVHSystem) findEntityCollisions(entities []ecs.Entity) {
	var wg sync.WaitGroup
	entitiesLength := len(entities)
	// get minimum 1 worker for small amount of entities, and maximum maxNumWorkers for a lot of entities
	numWorkers := max(min(entitiesLength/128, maxNumWorkers), 1)
	chunkSize := entitiesLength / numWorkers

	wg.Add(numWorkers)
	for workedId := 0; workedId < numWorkers; workedId++ {
		startIndex := workedId * chunkSize
		endIndex := startIndex + chunkSize - 1
		if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
			endIndex = entitiesLength
		}

		go func(start int, end int, id int) {
			defer wg.Done()

			for i := range entities[start:end] {
				entityA := entities[i+startIndex]

				potentialEntities := s.broadPhase(entityA, make([]ecs.Entity, 0, 64))
				if len(potentialEntities) == 0 {
					continue
				}
				s.narrowPhase(entityA, potentialEntities, id)
			}
		}(startIndex, endIndex, workedId)
	}
	// Wait for workers and close collision channel
	wg.Wait()
}

func (s *CollisionDetectionBVHSystem) registerCollisionEvents() {
	for i := range s.collisionEvents {
		events := &s.collisionEvents[i]
		events.EachData()(func(event *CollisionEvent) bool {
			pair := CollisionPair{event.entityA, event.entityB}
			s.currentCollisions[pair] = struct{}{}
			displacement := event.normal.Scale(event.depth)
			pos := event.position.Add(displacement)

			if _, exists := s.activeCollisions[pair]; !exists {
				proxy := s.EntityManager.Create()
				s.Collisions.Create(proxy, stdcomponents.Collision{
					E1:     pair.E1,
					E2:     pair.E2,
					State:  stdcomponents.CollisionStateEnter,
					Normal: event.normal,
					Depth:  event.depth,
				})

				s.Positions.Create(proxy, stdcomponents.Position{
					XY: vectors.Vec2{
						X: pos.X, Y: pos.Y,
					}})
				s.activeCollisions[pair] = proxy
			} else {
				proxy := s.activeCollisions[pair]
				collision := s.Collisions.GetUnsafe(proxy)
				position := s.Positions.GetUnsafe(proxy)
				collision.State = stdcomponents.CollisionStateStay
				collision.Depth = event.depth
				collision.Normal = event.normal
				position.XY.X = pos.X
				position.XY.Y = pos.Y
			}
			return true
		})
		events.Reset()
	}
}

func (s *CollisionDetectionBVHSystem) broadPhase(entityA ecs.Entity, result []ecs.Entity) []ecs.Entity {
	colliderA := s.GenericCollider.GetUnsafe(entityA)
	if colliderA.AllowSleep {
		isSleeping := s.ColliderSleepStateComponentManager.GetUnsafe(entityA)
		if isSleeping != nil {
			return result
		}
	}

	aabb := s.AABB.GetUnsafe(entityA)

	// Iterate through all trees
	for treeIndex := range s.trees {
		tree := &s.trees[treeIndex]
		layer := tree.Layer()

		// Check if mask includes this layer
		if !colliderA.Mask.HasLayer(layer) {
			continue
		}

		// Traverse this BVH tree for potential collisions
		result = tree.Query(aabb, result)
	}
	return result
}

func (s *CollisionDetectionBVHSystem) narrowPhase(entityA ecs.Entity, potentialEntities []ecs.Entity, workerId int) {
	for _, entityB := range potentialEntities {
		if entityA == entityB {
			continue
		}

		colliderA := s.GenericCollider.GetUnsafe(entityA)
		colliderB := s.GenericCollider.GetUnsafe(entityB)
		posA := s.Positions.GetUnsafe(entityA)
		posB := s.Positions.GetUnsafe(entityB)
		scaleA := s.Scales.GetUnsafe(entityA)
		scaleB := s.Scales.GetUnsafe(entityB)
		rotA := s.Rotations.GetUnsafe(entityA)
		rotB := s.Rotations.GetUnsafe(entityB)
		transformA := stdcomponents.Transform2d{
			Position: posA.XY,
			Rotation: rotA.Angle,
			Scale:    scaleA.XY,
		}
		transformB := stdcomponents.Transform2d{
			Position: posB.XY,
			Rotation: rotB.Angle,
			Scale:    scaleB.XY,
		}

		circleA := s.CircleColliders.GetUnsafe(entityA)
		circleB := s.CircleColliders.GetUnsafe(entityB)
		if circleA != nil && circleB != nil {
			radiusA := circleA.Radius * scaleA.XY.X
			radiusB := circleB.Radius * scaleB.XY.X
			if transformA.Position.Distance(transformB.Position) < radiusA+radiusB {
				s.collisionEvents[workerId].Append(CollisionEvent{
					entityA:  entityA,
					entityB:  entityB,
					position: transformA.Position,
					normal:   transformB.Position.Sub(transformA.Position).Normalize(),
					depth:    radiusA + radiusB - transformB.Position.Distance(transformA.Position),
				})
			}
			continue
		}

		// GJK strategy
		colA := s.getGjkCollider(colliderA, entityA)
		colB := s.getGjkCollider(colliderB, entityB)
		// First detect collision using GJK
		simplex, collision := gjk.CheckCollision(colA, colB, &transformA, &transformB)
		if !collision {
			continue
		}

		// If collision detected, get penetration details using EPA
		normal, depth := gjk.EPA(colA, colB, &transformA, &transformB, &simplex)
		position := posA.XY.Add(posB.XY.Sub(posA.XY))
		s.collisionEvents[workerId].Append(CollisionEvent{
			entityA:  entityA,
			entityB:  entityB,
			position: position,
			normal:   normal,
			depth:    depth,
		})
	}
}

func (s *CollisionDetectionBVHSystem) processExitStates() {
	for pair, proxy := range s.activeCollisions {
		if _, exists := s.currentCollisions[pair]; !exists {
			collision := s.Collisions.GetUnsafe(proxy)
			if collision.State == stdcomponents.CollisionStateExit {
				delete(s.activeCollisions, pair)
				s.EntityManager.Delete(proxy)
			} else {
				collision.State = stdcomponents.CollisionStateExit
			}
		}
	}
}

func (s *CollisionDetectionBVHSystem) getGjkCollider(collider *stdcomponents.GenericCollider, entity ecs.Entity) gjk.AnyCollider {
	switch collider.Shape {
	case stdcomponents.BoxColliderShape:
		return s.BoxColliders.GetUnsafe(entity)
	case stdcomponents.CircleColliderShape:
		return s.CircleColliders.GetUnsafe(entity)
	case stdcomponents.PolygonColliderShape:
		return s.PolygonColliders.GetUnsafe(entity)
	default:
		panic("unsupported collider shape")
	}
	return nil
}
