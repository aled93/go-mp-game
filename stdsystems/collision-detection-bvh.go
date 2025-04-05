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
	"image/color"
	"runtime"
	"sync"
	"time"
)

const debugTree = true

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
	s.GenericCollider.EachEntity(func(entity ecs.Entity) bool {
		aabb := s.AABB.Get(entity)
		layer := s.GenericCollider.Get(entity).Layer

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
			s.trees[i].Build()
			w.Done()
		}(i, wg)
	}
	wg.Wait()

	if debugTree {
		s.BvhTreeComponentManager.EachEntity(func(entity ecs.Entity) bool {
			s.EntityManager.Delete(entity)
			return true
		})

		for i := range s.trees {
			tree := s.trees[i]
			treeColor := color.RGBA{
				R: uint8(i * 255 / len(s.trees)),
				G: uint8((i + 1) * 255 / len(s.trees)),
				B: uint8((i + 2) * 255 / len(s.trees)),
				A: 30,
			}
			tree.AabbNodes.AllData(func(aabb *stdcomponents.AABB) bool {
				e := s.EntityManager.Create()
				s.BvhTreeComponentManager.Create(e, stdcomponents.BvhTree{
					Color: treeColor,
				})
				s.AABB.Create(e, *aabb)
				return true
			})
		}
	}

	if len(s.entities) < s.GenericCollider.Len() {
		s.entities = make([]ecs.Entity, 0, s.GenericCollider.Len())
	}
	s.entities = s.GenericCollider.RawEntities(s.entities)
	s.findEntityCollisions(s.entities)

	// could be used, but needs a worker id info
	//s.AABB.EachEntityParallel(func(entity ecs.Entity) bool {
	//
	//	potentialEntities := s.broadPhase(entity, make([]ecs.Entity, 0, 64))
	//	s.narrowPhase(entity, potentialEntities, id)
	//	return true
	//})

	for i := range s.collisionEvents {
		events := &s.collisionEvents[i]
		events.AllData(func(event *CollisionEvent) bool {
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
			return true
		})
		events.Reset()
	}
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

func (s *CollisionDetectionBVHSystem) broadPhase(entityA ecs.Entity, result []ecs.Entity) []ecs.Entity {
	colliderA := s.GenericCollider.Get(entityA)
	if colliderA.AllowSleep {
		isSleeping := s.ColliderSleepStateComponentManager.Get(entityA)
		if isSleeping != nil {
			return result
		}
	}

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
		result = tree.Query(aabb, result)
	}
	return result
}

func (s *CollisionDetectionBVHSystem) narrowPhase(entityA ecs.Entity, potentialEntities []ecs.Entity, workerId int) {
	for _, entityB := range potentialEntities {
		if entityA == entityB {
			continue
		}

		colliderA := s.GenericCollider.Get(entityA)
		colliderB := s.GenericCollider.Get(entityB)
		posA := s.Positions.Get(entityA)
		posB := s.Positions.Get(entityB)
		scaleA := s.Scales.Get(entityA)
		scaleB := s.Scales.Get(entityB)
		rotA := s.Rotations.Get(entityA)
		rotB := s.Rotations.Get(entityB)
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

		circleA := s.CircleColliders.Get(entityA)
		circleB := s.CircleColliders.Get(entityB)
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
