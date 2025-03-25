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

func NewCollisionDetectionBVHSystem() CollisionDetectionBVHSystem {
	return CollisionDetectionBVHSystem{
		activeCollisions: make(map[CollisionPair]ecs.Entity),
	}
}

type CollisionDetectionBVHSystem struct {
	EntityManager    *ecs.EntityManager
	Positions        *stdcomponents.PositionComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	GenericCollider  *stdcomponents.GenericColliderComponentManager
	BoxColliders     *stdcomponents.BoxColliderComponentManager
	CircleColliders  *stdcomponents.CircleColliderComponentManager
	PolygonColliders *stdcomponents.PolygonColliderComponentManager
	Collisions       *stdcomponents.CollisionComponentManager
	SpatialIndex     *stdcomponents.SpatialIndexComponentManager
	AABB             *stdcomponents.AABBComponentManager

	trees       []bvh.Tree2D
	treesLookup map[stdcomponents.CollisionLayer]int

	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
}

func (s *CollisionDetectionBVHSystem) Init() {}
func (s *CollisionDetectionBVHSystem) Run(dt time.Duration) {
	s.currentCollisions = make(map[CollisionPair]struct{})
	defer s.processExitStates()

	if s.AABB.Len() == 0 {
		return
	}

	// Init trees
	s.trees = make([]bvh.Tree2D, 0, 8)
	s.treesLookup = make(map[stdcomponents.CollisionLayer]int, 8)

	// Fill trees
	s.AABB.EachEntity(func(entity ecs.Entity) bool {
		aabb := s.AABB.Get(entity)
		layer := s.GenericCollider.Get(entity).Layer

		treeId, exists := s.treesLookup[layer]
		if !exists {
			treeId = len(s.trees)
			s.trees = append(s.trees, bvh.NewTree2D(layer, s.AABB.Len()))
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
			s.trees[i].Build()
			w.Done()
		}(i, wg)
	}
	wg.Wait()

	// Create collision channel
	collisionChan := make(chan CollisionEvent, 4096*4)
	doneChan := make(chan struct{})

	// Start result collector
	go func() {
		for event := range collisionChan {
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
				collision := s.Collisions.Get(proxy)
				position := s.Positions.Get(proxy)
				collision.State = stdcomponents.CollisionStateStay
				collision.Depth = event.depth
				collision.Normal = event.normal
				position.XY.X = pos.X
				position.XY.Y = pos.Y
			}
		}
		close(doneChan)
	}()

	entities := s.AABB.RawEntities(make([]ecs.Entity, 0, s.AABB.Len()))
	aabbs := s.AABB.RawComponents(make([]stdcomponents.AABB, 0, s.AABB.Len()))

	s.findEntityCollisions(entities, aabbs, collisionChan)

	close(collisionChan)
	<-doneChan // Wait for result collector
}

func (s *CollisionDetectionBVHSystem) Destroy() {}

func (s *CollisionDetectionBVHSystem) findEntityCollisions(entities []ecs.Entity, aabbs []stdcomponents.AABB, collisionChan chan<- CollisionEvent) {
	var wg sync.WaitGroup
	maxNumWorkers := runtime.NumCPU()
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

		go func(start int, end int) {
			defer wg.Done()

			for i := range entities[start:end] {
				entity := entities[i+startIndex]
				s.broadPhase(entity, collisionChan)
			}
		}(startIndex, endIndex)
	}
	// Wait for workers and close collision channel
	wg.Wait()
}

func (s *CollisionDetectionBVHSystem) broadPhase(entityA ecs.Entity, collisionChan chan<- CollisionEvent) {
	colliderA := s.GenericCollider.Get(entityA)
	aabb := s.AABB.Get(entityA)

	// Iterate through all trees
	for treeIndex := range s.trees {
		tree := &s.trees[treeIndex]
		layer := tree.Layer()

		// Check if mask includes this layer
		if !colliderA.Mask.HasLayer(layer) {
			continue
		}

		// Traverse this BVH tree for potential collisions
		tree.Query(*aabb, func(entityB ecs.Entity) {
			if entityA == entityB {
				return
			}

			colliderB := s.GenericCollider.Get(entityB)
			collision, ok := s.narrowPhase(colliderA, colliderB, entityA, entityB)
			if ok {
				collisionChan <- collision
			}
		})
	}
}

func (s *CollisionDetectionBVHSystem) narrowPhase(colliderA, colliderB *stdcomponents.GenericCollider, entityA, entityB ecs.Entity) (e CollisionEvent, ok bool) {
	colA := s.getGjkCollider(colliderA, entityA)
	colB := s.getGjkCollider(colliderB, entityB)
	posA := s.Positions.Get(entityA)
	posB := s.Positions.Get(entityB)
	scaleA := s.getScaleOrDefault(entityA)
	scaleB := s.getScaleOrDefault(entityB)
	rotA := s.getRotationOrDefault(entityA)
	rotB := s.getRotationOrDefault(entityB)
	transformA := stdcomponents.Transform2d{
		Position: posA.XY,
		Rotation: rotA.Angle,
		Scale:    scaleA,
	}
	transformB := stdcomponents.Transform2d{
		Position: posB.XY,
		Rotation: rotB.Angle,
		Scale:    scaleB,
	}

	// First detect collision using GJK
	simplex, collision := gjk.CheckCollision(colA, colB, &transformA, &transformB)
	if !collision {
		return e, false
	}

	// If collision detected, get penetration details using EPA
	normal, depth := gjk.EPA(colA, colB, &transformA, &transformB, &simplex)
	position := posA.XY.Add(posB.XY.Sub(posA.XY))
	return CollisionEvent{
		entityA:  entityA,
		entityB:  entityB,
		position: position,
		normal:   normal,
		depth:    depth,
	}, true
}

func (s *CollisionDetectionBVHSystem) getScaleOrDefault(entity ecs.Entity) vectors.Vec2 {
	scale := s.Scales.Get(entity)
	if scale != nil {
		return scale.XY // Dereference the component pointer
	}
	// Return default scale of 1 if component doesn't exist
	return vectors.Vec2{X: 1, Y: 1}
}

func (s *CollisionDetectionBVHSystem) getRotationOrDefault(entity ecs.Entity) stdcomponents.Rotation {
	rotation := s.Rotations.Get(entity)
	if rotation != nil {
		return *rotation // Dereference the component pointer
	}
	// Return default zero rotation if component doesn't exist
	return stdcomponents.Rotation{Angle: 0}
}

func (s *CollisionDetectionBVHSystem) processExitStates() {
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

func (s *CollisionDetectionBVHSystem) getGjkCollider(collider *stdcomponents.GenericCollider, entity ecs.Entity) gjk.AnyCollider {
	switch collider.Shape {
	case stdcomponents.BoxColliderShape:
		return s.BoxColliders.Get(entity)
	case stdcomponents.CircleColliderShape:
		return s.CircleColliders.Get(entity)
	case stdcomponents.PolygonColliderShape:
		return s.PolygonColliders.Get(entity)
	default:
		panic("unsupported collider shape")
	}
	return nil
}
