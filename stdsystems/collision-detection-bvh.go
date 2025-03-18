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
			pair := CollisionPair{event.entityA, event.entityB}
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

			if s.checkCollisionGjk(*colliderA, *colliderB, entityA, entityB) {
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

func (s *CollisionDetectionBVHSystem) checkCollisionGjk(colliderA, colliderB stdcomponents.GenericCollider, entityA, entityB ecs.Entity) bool {
	posA := s.Positions.Get(entityA)
	posB := s.Positions.Get(entityB)
	scaleA := s.getScaleOrDefault(entityA)
	scaleB := s.getScaleOrDefault(entityB)
	rotA := s.getRotationOrDefault(entityA) // Implement similar to getScaleOrDefault
	rotB := s.getRotationOrDefault(entityB)

	// Define support functions based on collider types
	supportA := s.getSupportFunction(entityA, colliderA, posA, &rotA, scaleA)
	supportB := s.getSupportFunction(entityB, colliderB, posB, &rotB, scaleB)

	return s.gjkCollides(supportA, supportB)
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
	scale := s.Scales.Get(entity)
	if scale != nil {
		return vectors.Vec2{X: scale.X, Y: scale.Y} // Dereference the component pointer
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

func (s *CollisionDetectionBVHSystem) getSupportFunction(entity ecs.Entity, collider stdcomponents.GenericCollider, pos *stdcomponents.Position, rot *stdcomponents.Rotation, scale vectors.Vec2) func(vectors.Vec2) vectors.Vec2 {
	switch collider.Shape {
	case stdcomponents.BoxColliderShape:
		box := s.BoxColliders.Get(entity)
		return func(d vectors.Vec2) vectors.Vec2 {
			return s.boxSupport(box, pos, rot, scale, d)
		}
	case stdcomponents.CircleColliderShape:
		circle := s.CircleColliders.Get(entity)
		return func(d vectors.Vec2) vectors.Vec2 {
			return s.circleSupport(circle, pos, scale, d)
		}
	case stdcomponents.PolygonColliderShape:
		poly := s.PolygonColliders.Get(entity)
		return func(d vectors.Vec2) vectors.Vec2 {
			return s.polygonSupport(poly, pos, rot, scale, d)
		}
	default:
		panic("unsupported collider shape")
	}
}

func (s *CollisionDetectionBVHSystem) circleSupport(circle *stdcomponents.CircleCollider, pos *stdcomponents.Position, scale vectors.Vec2, direction vectors.Vec2) vectors.Vec2 {
	if direction.LengthSquared() == 0 {
		return vectors.Vec2{X: pos.X, Y: pos.Y}
	}
	radius := circle.Radius * scale.X
	dirNorm := direction.Normalize()
	return vectors.Vec2{
		X: pos.X + dirNorm.X*radius,
		Y: pos.Y + dirNorm.Y*radius,
	}
}

func (s *CollisionDetectionBVHSystem) boxSupport(box *stdcomponents.BoxCollider, pos *stdcomponents.Position, rot *stdcomponents.Rotation, scale vectors.Vec2, direction vectors.Vec2) vectors.Vec2 {
	hw := (box.Width * scale.X) / 2
	hh := (box.Height * scale.Y) / 2

	// Rotate direction to local space
	localDir := direction.Rotate(-rot.Angle)

	localX := hw
	if localDir.X < 0 {
		localX = -hw
	}
	localY := hh
	if localDir.Y < 0 {
		localY = -hh
	}

	// Rotate back to world space and translate
	worldPoint := vectors.Vec2{X: localX, Y: localY}.Rotate(rot.Angle)
	return vectors.Vec2{
		X: pos.X + worldPoint.X,
		Y: pos.Y + worldPoint.Y,
	}
}

func (s *CollisionDetectionBVHSystem) polygonSupport(poly *stdcomponents.PolygonCollider, pos *stdcomponents.Position, rot *stdcomponents.Rotation, scale vectors.Vec2, direction vectors.Vec2) vectors.Vec2 {
	maxDot := math.Inf(-1)
	var maxVertex vectors.Vec2

	for _, v := range poly.Vertices {
		scaled := vectors.Vec2{X: v.X * scale.X, Y: v.Y * scale.Y}
		rotated := scaled.Rotate(rot.Angle)
		worldVertex := vectors.Vec2{X: pos.X + rotated.X, Y: pos.Y + rotated.Y}
		dot := float64(worldVertex.Dot(direction))
		if dot > maxDot {
			maxDot = dot
			maxVertex = worldVertex
		}
	}
	return maxVertex
}

func (s *CollisionDetectionBVHSystem) gjkCollides(supportA, supportB func(vectors.Vec2) vectors.Vec2) bool {
	direction := vectors.Vec2{X: 1, Y: 0} // Initial direction
	simplex := []vectors.Vec2{}
	for i := 0; i < 50; i++ { // Max iterations to prevent infinite loop
		p := s.minkowskiSupport(supportA, supportB, direction)
		if p.Dot(direction) < 0 {
			return false // No collision
		}
		simplex = append(simplex, p)
		if s.containsOrigin(simplex, direction) {
			return true
		}
	}
	return false
}

func (s *CollisionDetectionBVHSystem) minkowskiSupport(supportA, supportB func(vectors.Vec2) vectors.Vec2, d vectors.Vec2) vectors.Vec2 {
	a := supportA(d)
	b := supportB(d.Neg())
	return a.Sub(b)
}

func (s *CollisionDetectionBVHSystem) containsOrigin(simplex []vectors.Vec2, direction vectors.Vec2) bool {
	a := (simplex)[len(simplex)-1] // Last point added
	ao := a.Neg()                  // Vector from A to origin

	switch len(simplex) {
	case 3: // Triangle case
		b := (simplex)[1]
		c := (simplex)[0]

		ab := b.Sub(a)
		ac := c.Sub(a)

		// Perpendicular vectors
		abPerp := s.tripleProduct(ac, ab, ab)
		acPerp := s.tripleProduct(ab, ac, ac)

		// Region AB
		if abPerp.Dot(ao) > 0 {
			simplex = []vectors.Vec2{a, b}
			direction = abPerp
			return false
		}

		// Region AC
		if acPerp.Dot(ao) > 0 {
			simplex = []vectors.Vec2{a, c}
			direction = acPerp
			return false
		}

		// Inside triangle
		return true

	case 2: // Line segment case
		b := (simplex)[0]
		ab := b.Sub(a)

		// Perpendicular to AB facing origin
		abPerp := s.tripleProduct(ab, ao, ab)
		if abPerp.Dot(ao) > 0 {
			direction = abPerp
		} else {
			simplex = []vectors.Vec2{a}
			direction = ao
		}
		return false

	default:
		return false
	}
}

// Helper function for vector triple product
func (s *CollisionDetectionBVHSystem) tripleProduct(a, b, c vectors.Vec2) vectors.Vec2 {
	ac := a.Dot(c)
	bc := b.Dot(c)
	return vectors.Vec2{
		X: b.X*ac - a.X*bc,
		Y: b.Y*ac - a.Y*bc,
	}
}
