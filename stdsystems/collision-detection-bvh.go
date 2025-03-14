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
	"gomp/vectors"
	"math"
	"math/bits"
	"runtime"
	"slices"
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

	nodes []BVHNode

	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
}

type treeObject struct {
	Entity     ecs.Entity
	Bound      stdcomponents.AABB
	MortonCode uint32
}

func (s *CollisionDetectionBVHSystem) Init() {}
func (s *CollisionDetectionBVHSystem) Run(dt time.Duration) {
	if s.AABB.Len() == 0 {
		return
	}

	s.currentCollisions = make(map[CollisionPair]struct{})

	// Sorting by morton code
	treeObjects := make([]treeObject, 0, s.AABB.Len())
	s.AABB.EachEntity(func(entity ecs.Entity) bool {
		aabbBound := s.AABB.Get(entity)
		center := aabbBound.Min.Add(aabbBound.Max).Scale(0.5)
		newTreeObject := treeObject{
			Entity:     entity,
			Bound:      *aabbBound,
			MortonCode: morton2D(center.X, center.Y),
		}
		treeObjects = append(treeObjects, newTreeObject)
		return true
	})
	slices.SortFunc(treeObjects, func(a, b treeObject) int {
		return int(a.MortonCode) - int(b.MortonCode)
	})

	// extract data to the arrays
	entities := make([]ecs.Entity, len(treeObjects))
	aabbs := make([]stdcomponents.AABB, len(treeObjects))
	mortonCodes := make([]uint32, len(treeObjects))
	for i, obj := range treeObjects {
		entities[i] = obj.Entity
		aabbs[i] = obj.Bound
		mortonCodes[i] = obj.MortonCode
	}

	s.buildBVH(entities, aabbs, mortonCodes)

	// Create collision channel
	collisionChan := make(chan CollisionEvent, 4096*runtime.NumCPU())
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

	s.findEntityCollisions(entities, aabbs, collisionChan)

	close(collisionChan)
	<-doneChan // Wait for result collector
	s.processExitStates()
}
func (s *CollisionDetectionBVHSystem) Destroy() {}

func (s *CollisionDetectionBVHSystem) findEntityCollisions(entities []ecs.Entity, aabbs []stdcomponents.AABB, collisionChan chan<- CollisionEvent) {
	var wg sync.WaitGroup
	maxNumWorkers := runtime.NumCPU()
	entitiesLength := len(entities)
	// get minimum 1 worker for small amount of entities, and maximum maxNumWorkers for a lot of entities
	numWorkers := max(min(entitiesLength/32, maxNumWorkers), 1)
	chunkSize := entitiesLength / numWorkers

	wg.Add(numWorkers)
	for workedId := 0; workedId < numWorkers; workedId++ {
		startIndex := workedId * chunkSize
		endIndex := startIndex + chunkSize - 1
		if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
			endIndex = entitiesLength
		}
		rootIndex := len(s.nodes) - 1

		go func(start int, end int) {
			defer wg.Done()

			for i := range entities[start:end] {
				s.traverseBVHForCollisions(entities, aabbs, i+startIndex, rootIndex, collisionChan)
			}
		}(startIndex, endIndex)
	}
	// Wait for workers and close collision channel
	wg.Wait()

}

func (s *CollisionDetectionBVHSystem) traverseBVHForCollisions(entities []ecs.Entity, aabbs []stdcomponents.AABB, i int, nodeIndex int, collisionChan chan<- CollisionEvent) {
	node := s.nodes[nodeIndex]
	entityA := entities[i]
	aabbA := aabbs[i]

	if node.Left == -1 && node.Right == -1 {
		j := nodeIndex // Leaf nodes are ordered as per entities slice
		if j > i {
			entityB := entities[j]
			if aabbOverlap(aabbA, node.Bounds) {
				colliderA := s.GenericCollider.Get(entityA)
				colliderB := s.GenericCollider.Get(entityB)

				if colliderA.Mask&(1<<colliderB.Layer) == 0 &&
					colliderB.Mask&(1<<colliderA.Layer) == 0 {
					return
				}

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
			}
		}
		return
	}

	if node.Left != -1 {
		leftNode := s.nodes[node.Left]
		if aabbOverlap(aabbA, leftNode.Bounds) {
			s.traverseBVHForCollisions(entities, aabbs, i, node.Left, collisionChan)
		}
	}
	if node.Right != -1 {
		rightNode := s.nodes[node.Right]
		if aabbOverlap(aabbA, rightNode.Bounds) {
			s.traverseBVHForCollisions(entities, aabbs, i, node.Right, collisionChan)
		}
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
	s.nodes = make([]BVHNode, 0, len(entities)*2)

	// Create leaf nodes in morton order
	leaves := make([]BVHNode, len(entities))
	for i := range entities {
		leaves[i] = BVHNode{
			Entity: entities[i],
			Bounds: aabbs[i],
			Left:   -1,
			Right:  -1,
		}
	}

	// Build hierarchy using morton codes
	s.nodes = append(s.nodes, leaves...)
	s.buildHierarchy(mortonCodes, 0, len(leaves)-1)
}

// buildHierarchy recursively constructs BVH using morton codes
func (s *CollisionDetectionBVHSystem) buildHierarchy(mortonCodes []uint32, start, end int) int {
	if start == end {
		return start // Leaf node
	}

	// Find split point using the highest differing bit
	split := findSplit(mortonCodes, start, end)

	// Recursively build left and right subtrees
	left := s.buildHierarchy(mortonCodes, start, split)
	right := s.buildHierarchy(mortonCodes, split+1, end)

	// Create internal node
	node := BVHNode{
		Left:   left,
		Right:  right,
		Bounds: mergeAABB(s.nodes[left].Bounds, s.nodes[right].Bounds),
	}

	s.nodes = append(s.nodes, node)
	return len(s.nodes) - 1
}

// findSplit finds the position where the highest bit changes
func findSplit(sortedMortonCodes []uint32, start, end int) int {
	// Identical Morton sortedMortonCodes => split the range in the middle.
	first := sortedMortonCodes[start]
	last := sortedMortonCodes[end]

	if first == last {
		return (start + end) >> 1
	}

	// Calculate the number of highest bits that are the same
	// for all objects, using the count-leading-zeros intrinsic.
	commonPrefix := bits.LeadingZeros32(first ^ last)

	// Use binary search to find where the next bit differs.
	// Specifically, we are looking for the highest object that
	// shares more than commonPrefix bits with the first one.
	split := start
	step := end - start

	for {
		step = (step + 1) >> 1   // exponential decrease
		newSplit := split + step // proposed new position

		if newSplit < end {
			splitCode := sortedMortonCodes[newSplit]
			splitPrefix := bits.LeadingZeros32(first ^ splitCode)
			if splitPrefix > commonPrefix {
				split = newSplit
			}
		}

		if step <= 1 {
			break
		}
	}

	return split
}

// aabbOverlap checks if two AABBs intersect
func aabbOverlap(a, b stdcomponents.AABB) bool {
	return a.Max.X >= b.Min.X && a.Min.X <= b.Max.X &&
		a.Max.Y >= b.Min.Y && a.Min.Y <= b.Max.Y
}

// mergeAABB combines two AABBs
func mergeAABB(a, b stdcomponents.AABB) stdcomponents.AABB {
	return stdcomponents.AABB{
		Min: vectors.Vec2{
			X: min(a.Min.X, b.Min.X),
			Y: min(a.Min.Y, b.Min.Y),
		},
		Max: vectors.Vec2{
			X: max(a.Max.X, b.Max.X),
			Y: max(a.Max.Y, b.Max.Y),
		},
	}
}

// Expands a 10-bit integer into 20 bits by inserting 1 zero after each bit
func expandBits2D(v uint32) uint32 {
	v = (v * 0x00010001) & 0xFF0000FF
	v = (v * 0x00000101) & 0x0F00F00F
	v = (v * 0x00000011) & 0xC30C30C3
	v = (v * 0x00000005) & 0x24924924
	return v
}

// 2D Morton code for coordinates in [0,1] range
func morton2D(x, y float32) uint32 {
	xx := uint32(math.Min(math.Max(float64(x)*1024.0, 0.0), 1023.0))
	yy := uint32(math.Min(math.Max(float64(y)*1024.0, 0.0), 1023.0))
	return (expandBits2D(xx) << 1) | expandBits2D(yy)
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
