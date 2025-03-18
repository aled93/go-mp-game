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

// Node represents a node in the BVH tree
type Node struct {
	Left, Right int // Child indices, -1 for leaf
	MortonCode  uint32
	Entity      ecs.Entity
	Bounds      stdcomponents.AABB
}

func (n *Node) isLeaf() bool {
	return n.Left == -1 && n.Right == -1
}

func NewTree2D(layer stdcomponents.CollisionLayer, prealloc int) Tree2D {
	return Tree2D{
		layer:      layer,
		components: make([]treeComponent, 0, prealloc),
	}
}

// Tree2D represents a BVH tree
type Tree2D struct {
	layer      stdcomponents.CollisionLayer
	nodes      []Node
	components []treeComponent
	rootIndex  int
}

type treeComponent struct {
	Entity ecs.Entity
	AABB   stdcomponents.AABB
}

func (t *Tree2D) Layer() stdcomponents.CollisionLayer {
	return t.layer
}

func (t *Tree2D) Query(aabb stdcomponents.AABB, handler func(entity ecs.Entity)) {
	if t.rootIndex == -1 || len(t.nodes) == 0 { // Handle empty tree
		return
	}

	// Use stack-based traversal
	stack := make([]int, 0, 64)
	stack = append(stack, t.rootIndex)

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := t.nodes[nodeIndex]

		// Early exit if no AABB overlap
		if !t.aabbOverlap(&aabb, &node.Bounds) {
			continue
		}

		if node.isLeaf() {
			// Check detailed collision with node.Entity
			// (Same as your existing collision logic)
			handler(node.Entity)
		} else {
			// Push children to stack
			stack = append(stack, node.Right, node.Left)
		}
	}
}

func (t *Tree2D) AddComponent(entity ecs.Entity, aabbs stdcomponents.AABB) {
	t.components = append(t.components, treeComponent{
		Entity: entity,
		AABB:   aabbs,
	})
}

type TaskType int

const (
	BuildTaskType TaskType = iota
	MergeTaskType
)

type Task struct {
	Type  TaskType
	Start int
	End   int
}

func (t *Tree2D) Build() {
	if cap(t.nodes) < len(t.components)*2 {
		t.nodes = make([]Node, 0, len(t.components)*2)
	}
	t.nodes = t.nodes[:0]

	// Create leaf nodes
	leaves := make([]Node, len(t.components))
	for i := range t.components {
		aabb := t.components[i].AABB
		center := aabb.Min.Add(aabb.Max).Scale(0.5)
		code := t.morton2D(center.X, center.Y)
		leaves[i] = Node{
			Left:       -1,
			Right:      -1,
			MortonCode: code,
			Entity:     t.components[i].Entity,
			Bounds:     aabb,
		}
	}
	t.components = t.components[:0]

	// Sort leaf nodes by morton code
	slices.SortFunc(leaves, func(a, b Node) int {
		return int(a.MortonCode) - int(b.MortonCode)
	})

	t.nodes = append(t.nodes, leaves...)

	// Stack-based hierarchy construction
	var resultStack []int
	taskStack := []Task{{Type: BuildTaskType, Start: 0, End: len(leaves) - 1}}

	for len(taskStack) > 0 {
		task := taskStack[len(taskStack)-1]
		taskStack = taskStack[:len(taskStack)-1]

		switch task.Type {
		case BuildTaskType:
			start, end := task.Start, task.End
			if start == end {
				// Leaf node: push its index to result stack
				resultStack = append(resultStack, start)
			} else {
				split := t.findSplit(start, end)
				// Schedule MergeTask after processing children
				taskStack = append(taskStack, Task{Type: MergeTaskType})
				// Process right child first (LIFO order)
				taskStack = append(taskStack, Task{Type: BuildTaskType, Start: split + 1, End: end})
				// Process left child next
				taskStack = append(taskStack, Task{Type: BuildTaskType, Start: start, End: split})
			}
		case MergeTaskType:
			// Pop right then left from result stack
			right := resultStack[len(resultStack)-1]
			resultStack = resultStack[:len(resultStack)-1]
			left := resultStack[len(resultStack)-1]
			resultStack = resultStack[:len(resultStack)-1]

			// Create parent node and append to t.nodes
			parent := Node{
				Left:   left,
				Right:  right,
				Bounds: t.mergeAABB(&t.nodes[left].Bounds, &t.nodes[right].Bounds),
			}
			t.nodes = append(t.nodes, parent)
			// Push parent index to result stack
			resultStack = append(resultStack, len(t.nodes)-1)
		}
	}

	// After processing all tasks, resultStack holds root index
	if len(resultStack) > 0 {
		t.rootIndex = resultStack[0]
	}
}

// findSplit finds the position where the highest bit changes
func (t *Tree2D) findSplit(start, end int) int {
	// Identical Morton sortedMortonCodes => split the range in the middle.
	first := t.nodes[start].MortonCode
	last := t.nodes[end].MortonCode

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
			splitCode := t.nodes[newSplit].MortonCode
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
func (t *Tree2D) mergeAABB(a, b *stdcomponents.AABB) stdcomponents.AABB {
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

// go:inline aabbOverlap checks if two AABB intersect
func (t *Tree2D) aabbOverlap(a, b *stdcomponents.AABB) bool {
	// Check for non-overlap conditions first (early exit)
	if a.Max.X < b.Min.X || a.Min.X > b.Max.X {
		return false
	}
	if a.Max.Y < b.Min.Y || a.Min.Y > b.Max.Y {
		return false
	}
	return true
}

// Expands a 10-bit integer into 20 bits by inserting 1 zero after each bit
func (t *Tree2D) expandBits2D(v uint32) uint32 {
	v = (v * 0x00010001) & 0xFF0000FF
	v = (v * 0x00000101) & 0x0F00F00F
	v = (v * 0x00000011) & 0xC30C30C3
	v = (v * 0x00000005) & 0x24924924
	return v
}

// 2D Morton code for centroids coordinates in [0,1] range
func (t *Tree2D) morton2D(x, y float32) uint32 {
	xx := uint32(math.Min(math.Max(float64(x)*1024.0, 0.0), 1023.0))
	yy := uint32(math.Min(math.Max(float64(y)*1024.0, 0.0), 1023.0))
	return (t.expandBits2D(xx) << 1) | t.expandBits2D(yy)
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
