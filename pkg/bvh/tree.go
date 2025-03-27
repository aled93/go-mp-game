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

package bvh

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
	"math/bits"
	"slices"
)

type node struct {
	childIndex int32 // if < 0 then points to leaf
}

type leaf struct {
	id ecs.Entity
}

type component struct {
	entity ecs.Entity
	aabb   *stdcomponents.AABB
	code   uint32
}

func NewGenTree(layer stdcomponents.CollisionLayer, prealloc int) Tree {
	return Tree{
		nodes:      make([]node, 0, prealloc),
		aabbNodes:  make([]stdcomponents.AABB, 0, prealloc),
		leaves:     make([]leaf, 0, prealloc),
		aabbLeaves: make([]*stdcomponents.AABB, 0, prealloc),
		codes:      make([]uint32, 0, prealloc),
		components: make([]component, 0, prealloc),
		layer:      layer,
	}
}

type Tree struct {
	nodes     []node
	aabbNodes []stdcomponents.AABB

	leaves     []leaf
	aabbLeaves []*stdcomponents.AABB

	codes []uint32

	components []component

	layer stdcomponents.CollisionLayer
}

func (t *Tree) AddComponent(entity ecs.Entity, aabb *stdcomponents.AABB) {
	center := aabb.Min.Add(aabb.Max).Scale(0.5)
	code := t.morton2D(center.X, center.Y)
	t.components = append(t.components, component{
		entity: entity,
		aabb:   aabb,
		code:   code,
	})
}

func (t *Tree) Build() {
	// Reset tree
	t.nodes = t.nodes[:0]
	t.aabbNodes = t.aabbNodes[:0]
	t.leaves = t.leaves[:0]
	t.aabbLeaves = t.aabbLeaves[:0]

	// Sort components by morton code
	slices.SortFunc(t.components, func(a, b component) int {
		return int(a.code) - int(b.code)
	})

	// Add leaves
	for i := 0; i < len(t.components); i++ {
		component := &t.components[i]
		t.leaves = append(t.leaves, leaf{
			id: component.entity,
		})
		t.aabbLeaves = append(t.aabbLeaves, component.aabb)
		t.codes = append(t.codes, component.code)
	}
	t.components = t.components[:0]

	// Add root node
	t.nodes = append(t.nodes, node{-1})
	t.aabbNodes = append(t.aabbNodes, stdcomponents.AABB{})

	t.buildH(0, 0, len(t.leaves)-1)
}

func (t *Tree) buildH(parentIndex int, start, end int) {
	if start == end {
		// Is a leaf
		t.nodes[parentIndex].childIndex = -int32(start)
		t.aabbNodes[parentIndex] = *t.aabbLeaves[start]
		return
	}

	split := t.findSplit(start, end)

	// Add left node
	leftIndex := len(t.nodes)
	t.nodes = append(t.nodes, node{-1})
	t.aabbNodes = append(t.aabbNodes, stdcomponents.AABB{})

	// Add right node
	rightIndex := len(t.nodes)
	t.nodes = append(t.nodes, node{-1})
	t.aabbNodes = append(t.aabbNodes, stdcomponents.AABB{})

	t.nodes[parentIndex].childIndex = int32(leftIndex)

	t.buildH(leftIndex, start, split)
	t.buildH(rightIndex, split+1, end)

	t.aabbNodes[parentIndex] = t.mergeAABB(&t.aabbNodes[leftIndex], &t.aabbNodes[rightIndex])
}

func (t *Tree) Layer() stdcomponents.CollisionLayer {
	return t.layer
}

func (t *Tree) Query(aabb *stdcomponents.AABB, result []ecs.Entity) []ecs.Entity {
	if len(t.nodes) == 0 { // Handle empty tree
		return result
	}

	// Use stack-based traversal
	const stackSize = 32
	stack := [stackSize]int{}
	stackPtr := 0
	stack[stackPtr] = 0
	stackPtr++

	for stackPtr > 0 {
		stackPtr--
		nodeIndex := stack[stackPtr]
		a := &t.aabbNodes[nodeIndex]
		b := aabb

		// Early exit if no AABB overlap
		if !t.aabbOverlap(a, b) {
			continue
		}

		node := &t.nodes[nodeIndex]
		if node.childIndex <= 0 {
			// Is a leaf
			index := -node.childIndex
			result = append(result, t.leaves[index].id)
			continue
		}

		// Push child indices (right and left) onto the stack.
		stack[stackPtr] = int(node.childIndex + 1)
		stack[stackPtr+1] = int(node.childIndex)
		stackPtr += 2
	}

	return result
}

// go:inline aabbOverlap checks if two AABB intersect
func (t *Tree) aabbOverlap(a, b *stdcomponents.AABB) bool {
	return a.Max.X >= b.Min.X && a.Min.X <= b.Max.X &&
		a.Max.Y >= b.Min.Y && a.Min.Y <= b.Max.Y
}

// findSplit finds the position where the highest bit changes
func (t *Tree) findSplit(start, end int) int {
	// Identical Morton sortedMortonCodes => split the range in the middle.
	first := t.codes[start]
	last := t.codes[end]

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
			splitCode := t.codes[newSplit]
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

// mergeAABB combines two AABB
func (t *Tree) mergeAABB(a, b *stdcomponents.AABB) stdcomponents.AABB {
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
func (t *Tree) expandBits2D(v uint32) uint32 {
	v = (v | (v << 16)) & 0x030000FF
	v = (v | (v << 8)) & 0x0300F00F
	v = (v | (v << 4)) & 0x030C30C3
	v = (v | (v << 2)) & 0x09249249
	return v
}

const mortonPrecision = 1 << 16

// 2D Morton code for centroids coordinates in [0,1] range
func (t *Tree) morton2D(x, y float32) uint32 {
	xx := uint32(math.Min(math.Max(float64(x)*mortonPrecision, 0.0), mortonPrecision-1))
	yy := uint32(math.Min(math.Max(float64(y)*mortonPrecision, 0.0), mortonPrecision-1))
	return (t.expandBits2D(xx) << 1) | t.expandBits2D(yy)
}
