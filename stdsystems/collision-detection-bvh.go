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
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
	"runtime"
	"sync"
	"time"
)

// BVHNode represents a node in the BVH tree
type BVHNode struct {
	Left, Right int // Child indices, -1 for leaf
	Entity      ecs.Entity
	Bounds      stdcomponents.AABB
}

func NewCollisionDetectionBVHSystem() CollisionDetectionBVHSystem {
	return CollisionDetectionBVHSystem{
		activeCollisions: make(map[CollisionPair]ecs.Entity),
	}
}

type CollisionDetectionBVHSystem struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	Scales          *stdcomponents.ScaleComponentManager
	GenericCollider *stdcomponents.GenericColliderComponentManager
	BoxColliders    *stdcomponents.BoxColliderComponentManager
	CircleColliders *stdcomponents.CircleColliderComponentManager
	Collisions      *stdcomponents.CollisionComponentManager
	SpatialIndex    *stdcomponents.SpatialIndexComponentManager
	AABB            *stdcomponents.AABBComponentManager

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

	s.trees = make([]bvh.Tree2D, 0, 8)
	s.treesLookup = make(map[stdcomponents.CollisionLayer]int, 8)

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
				s.checkEntityCollisions(entity, collisionChan)
			}
		}(startIndex, endIndex)
	}
	// Wait for workers and close collision channel
	wg.Wait()
}

func (s *CollisionDetectionBVHSystem) checkEntityCollisions(entityA ecs.Entity, collisionChan chan<- CollisionEvent) {
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

			if s.checkCollision(*colliderA, *colliderB, entityA, entityB) {
				posA := s.Positions.Get(entityA)
				posB := s.Positions.Get(entityB)
				posX := (posA.X + posB.X) / 2
				posY := (posA.Y + posB.Y) / 2
				collisionChan <- CollisionEvent{
					entityA: entityA,
					entityB: entityB,
					posX:    posX,
					posY:    posY,
				}
			}
		})
	}
}

func (s *CollisionDetectionBVHSystem) checkCollision(colliderA, colliderB stdcomponents.GenericCollider, entityA, entityB ecs.Entity) bool {
	posA := s.Positions.Get(entityA)
	posB := s.Positions.Get(entityB)
	scaleA := s.getScaleOrDefault(entityA)
	scaleB := s.getScaleOrDefault(entityB)

	switch colliderA.Shape {
	case stdcomponents.BoxColliderShape:
		a := s.BoxColliders.Get(entityA)
		switch colliderB.Shape {
		case stdcomponents.BoxColliderShape:
			return true // AABB overlap already confirmed
		case stdcomponents.CircleColliderShape:
			b := s.CircleColliders.Get(entityB)
			return s.circleVsBox(b, *posB, scaleB, a, *posA, scaleA)
		default:
			return false
		}
	case stdcomponents.CircleColliderShape:
		a := s.CircleColliders.Get(entityA)
		switch colliderB.Shape {
		case stdcomponents.BoxColliderShape:
			b := s.BoxColliders.Get(entityB)
			return s.circleVsBox(a, *posA, scaleA, b, *posB, scaleB)
		case stdcomponents.CircleColliderShape:
			b := s.CircleColliders.Get(entityB)
			dx := posA.X - posB.X
			dy := posA.Y - posB.Y
			distanceSq := dx*dx + dy*dy
			radiusA := a.Radius * scaleA.X
			radiusB := b.Radius * scaleB.X
			return distanceSq <= (radiusA+radiusB)*(radiusA+radiusB)
		default:
			return false
		}
	default:
		return false
	}
}

func (s *CollisionDetectionBVHSystem) getScaleOrDefault(entity ecs.Entity) vectors.Vec2 {
	if s.Scales.Has(entity) {
		scale := s.Scales.Get(entity)
		return vectors.Vec2{X: scale.X, Y: scale.Y}
	}
	return vectors.Vec2{X: 1, Y: 1}
}

func (s *CollisionDetectionBVHSystem) circleVsBox(circleCollider *stdcomponents.CircleCollider, circlePos stdcomponents.Position, circleScale vectors.Vec2, boxCollider *stdcomponents.BoxCollider, boxPos stdcomponents.Position, boxScale vectors.Vec2) bool {
	radius := circleCollider.Radius * circleScale.X
	boxWidth := boxCollider.Width * boxScale.X
	boxHeight := boxCollider.Height * boxScale.Y

	halfWidth := boxWidth / 2
	halfHeight := boxHeight / 2

	boxMinX := boxPos.X - halfWidth
	boxMaxX := boxPos.X + halfWidth
	boxMinY := boxPos.Y - halfHeight
	boxMaxY := boxPos.Y + halfHeight

	closestX := max(boxMinX, min(circlePos.X, boxMaxX))
	closestY := max(boxMinY, min(circlePos.Y, boxMaxY))

	dx := circlePos.X - closestX
	dy := circlePos.Y - closestY
	distanceSq := dx*dx + dy*dy

	return distanceSq <= radius*radius
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

// buildBVH constructs hierarchy using sorted morton codes
func (s *CollisionDetectionBVHSystem) buildBVH(entities []ecs.Entity, aabbs []stdcomponents.AABB, mortonCodes []uint32) {

}

// Expands a 10-bit integer into 30 bits by inserting 2 zeros after each bit
func expandBits3D(v uint32) uint32 {
	v = (v * 0x00010001) & 0xFF0000FF
	v = (v * 0x00000101) & 0x0F00F00F
	v = (v * 0x00000011) & 0xC30C30C3
	v = (v * 0x00000005) & 0x49249249
	return v
}

// 3D Morton code for coordinates in [0,1] range
func morton3D(x, y, z float32) uint32 {
	xx := uint32(math.Min(math.Max(float64(x)*1024.0, 0.0), 1023.0))
	yy := uint32(math.Min(math.Max(float64(y)*1024.0, 0.0), 1023.0))
	zz := uint32(math.Min(math.Max(float64(z)*1024.0, 0.0), 1023.0))
	return expandBits3D(xx)*4 + expandBits3D(yy)*2 + expandBits3D(zz)
}
