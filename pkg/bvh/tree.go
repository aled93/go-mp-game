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

func NewTree(layer stdcomponents.CollisionLayer) Tree {
	return Tree{
		nodes:      ecs.NewPagedArray[node](),
		aabbNodes:  ecs.NewPagedArray[stdcomponents.AABB](),
		leaves:     ecs.NewPagedArray[leaf](),
		aabbLeaves: ecs.NewPagedArray[*stdcomponents.AABB](),
		codes:      ecs.NewPagedArray[uint32](),
		components: ecs.NewPagedArray[component](),
		layer:      layer,
	}
}

type Tree struct {
	nodes      ecs.PagedArray[node]
	aabbNodes  ecs.PagedArray[stdcomponents.AABB]
	leaves     ecs.PagedArray[leaf]
	aabbLeaves ecs.PagedArray[*stdcomponents.AABB]
	codes      ecs.PagedArray[uint32]
	components ecs.PagedArray[component]
	layer      stdcomponents.CollisionLayer
}

func (t *Tree) AddComponent(entity ecs.Entity, aabb *stdcomponents.AABB) {
	center := aabb.Min.Add(aabb.Max).Scale(0.5)
	code := t.morton2D(center.X, center.Y)
	t.components.Append(component{
		entity: entity,
		aabb:   aabb,
		code:   code,
	})
}

func (t *Tree) Build() {
	// Reset tree
	t.nodes.Reset()
	t.aabbNodes.Reset()
	t.leaves.Reset()
	t.aabbLeaves.Reset()
	t.codes.Reset()

	// Extract and sort components by morton code
	var componentsSlice = make([]component, 0, t.components.Len())
	for i := 0; i < t.components.Len(); i++ {
		componentsSlice = append(componentsSlice, t.components.GetValue(i))
	}
	slices.SortFunc(componentsSlice, func(a, b component) int {
		return int(a.code) - int(b.code)
	})

	// Add leaves
	for i := range componentsSlice {
		component := &componentsSlice[i]
		t.leaves.Append(leaf{id: component.entity})
		t.aabbLeaves.Append(component.aabb)
		t.codes.Append(component.code)
	}
	t.components.Reset()

	if t.leaves.Len() == 0 {
		return
	}

	// Add root node
	t.nodes.Append(node{-1})
	t.aabbNodes.Append(stdcomponents.AABB{})

	type buildTask struct {
		parentIndex     int
		start           int
		end             int
		childrenCreated bool
	}

	stack := []buildTask{
		{parentIndex: 0, start: 0, end: t.leaves.Len() - 1, childrenCreated: false},
	}

	for len(stack) > 0 {
		// Pop the last task
		task := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !task.childrenCreated {
			if task.start == task.end {
				// Leaf node
				t.nodes.Get(task.parentIndex).childIndex = -int32(task.start)
				t.aabbNodes.Set(task.parentIndex, *t.aabbLeaves.GetValue(task.start))
				continue
			}

			split := t.findSplit(task.start, task.end)

			// Create left and right nodes
			leftIndex := t.nodes.Len()
			t.nodes.Append(node{-1})
			t.nodes.Append(node{-1})
			t.aabbNodes.Append(stdcomponents.AABB{})
			t.aabbNodes.Append(stdcomponents.AABB{})

			// Set parent's childIndex to leftIndex
			t.nodes.Get(task.parentIndex).childIndex = int32(leftIndex)

			// Push parent task back with childrenCreated=true
			stack = append(stack, buildTask{
				parentIndex:     task.parentIndex,
				start:           task.start,
				end:             task.end,
				childrenCreated: true,
			})

			// Push right child task (split+1 to end)
			stack = append(stack, buildTask{
				parentIndex:     leftIndex + 1,
				start:           split + 1,
				end:             task.end,
				childrenCreated: false,
			})

			// Push left child task (start to split)
			stack = append(stack, buildTask{
				parentIndex:     leftIndex,
				start:           task.start,
				end:             split,
				childrenCreated: false,
			})
		} else {
			// Merge children's AABBs into parent
			leftChildIndex := int(t.nodes.Get(task.parentIndex).childIndex)
			rightChildIndex := leftChildIndex + 1

			leftAABB := t.aabbNodes.Get(leftChildIndex)
			rightAABB := t.aabbNodes.Get(rightChildIndex)

			merged := t.mergeAABB(leftAABB, rightAABB)
			t.aabbNodes.Set(task.parentIndex, merged)
		}
	}
}

func (t *Tree) Layer() stdcomponents.CollisionLayer {
	return t.layer
}

func (t *Tree) Query(aabb *stdcomponents.AABB, result []ecs.Entity) []ecs.Entity {
	if t.nodes.Len() == 0 { // Handle empty tree
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
		a := t.aabbNodes.Get(nodeIndex)
		b := aabb

		// Early exit if no AABB overlap
		if !t.aabbOverlap(a, b) {
			continue
		}

		node := t.nodes.Get(nodeIndex)
		if node.childIndex <= 0 {
			// Is a leaf
			index := -int(node.childIndex)
			result = append(result, t.leaves.Get(index).id)
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
	first := t.codes.GetValue(start)
	last := t.codes.GetValue(end)

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
			splitCode := t.codes.GetValue(newSplit)
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

// Expands a 16-bit integer into 32 bits by inserting 1 zero after each bit
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
