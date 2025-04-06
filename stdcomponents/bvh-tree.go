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

package stdcomponents

import (
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"math"
)

const mortonPrecision = (1 << 16) - 1

type BvhNode struct {
	ChildIndex int32 // if < 0 then points to BvhLeaf
}

type BvhLeaf struct {
	Id ecs.Entity
}

type BvhComponent struct {
	Entity ecs.Entity
	Aabb   AABB
	Code   uint64
}

type BvhTree struct {
	Nodes      ecs.PagedArray[BvhNode]
	AabbNodes  ecs.PagedArray[AABB]
	Leaves     ecs.PagedArray[BvhLeaf]
	AabbLeaves ecs.PagedArray[AABB]
	Codes      ecs.PagedArray[uint64]
	Components ecs.PagedArray[BvhComponent]

	ComponentsSlice []BvhComponent
}

func (t *BvhTree) AddComponent(entity ecs.Entity, aabb AABB) {
	code := t.morton2D(&aabb)
	t.Components.Append(BvhComponent{
		Entity: entity,
		Aabb:   aabb,
		Code:   code,
	})
}

func (t *BvhTree) Query(aabb *AABB, result []ecs.Entity) []ecs.Entity {
	if t.Nodes.Len() == 0 { // Handle empty tree
		return result
	}

	// Use stack-based traversal
	const stackSize = 32
	stack := [stackSize]int32{0}
	stackLen := 1

	for stackLen > 0 {
		stackLen--
		nodeIndex := int(stack[stackLen])
		a := t.AabbNodes.Get(nodeIndex)
		b := aabb

		// Early exit if no AABB overlap
		if !t.aabbOverlap(a, b) {
			continue
		}

		node := t.Nodes.Get(nodeIndex)
		if node.ChildIndex <= 0 {
			// Is a leaf
			index := -int(node.ChildIndex)
			leafAabb := t.AabbLeaves.Get(index)
			if t.aabbOverlap(leafAabb, aabb) {
				result = append(result, t.Leaves.Get(index).Id)
			}
			continue
		}

		// Push child indices (right and left) onto the stack.
		stack[stackLen] = node.ChildIndex + 1
		stack[stackLen+1] = node.ChildIndex
		stackLen += 2
	}

	return result
}

// go:inline aabbOverlap checks if two AABB intersect
func (t *BvhTree) aabbOverlap(a, b *AABB) bool {
	return a.Max.X >= b.Min.X && a.Min.X <= b.Max.X &&
		a.Max.Y >= b.Min.Y && a.Min.Y <= b.Max.Y
}

// Expands a 16-bit integer into 32 bits by inserting 1 zero after each bit
func (t *BvhTree) expandBits2D(v uint32) uint32 {
	v = (v | (v << 8)) & 0x00FF00FF
	v = (v | (v << 4)) & 0x0F0F0F0F
	v = (v | (v << 2)) & 0x33333333
	v = (v | (v << 1)) & 0x55555555
	return v
}

func (t *BvhTree) morton2D(aabb *AABB) uint64 {
	center := aabb.Center()
	// Scale coordinates to 16-bit integers
	//assert.True(center.X >= 0 && center.Y >= 0, "morton2D: center out of range")

	xx := uint64(float64(center.X) * mortonPrecision)
	yy := uint64(float64(center.Y) * mortonPrecision)

	assert.True(xx < math.MaxUint64, "morton2D: x out of range")
	assert.True(yy < math.MaxUint64, "morton2D: y out of range")

	// Spread the bits of x into the even positions
	xx = (xx | (xx << 16)) & 0x0000FFFF0000FFFF
	xx = (xx | (xx << 8)) & 0x00FF00FF00FF00FF
	xx = (xx | (xx << 4)) & 0x0F0F0F0F0F0F0F0F
	xx = (xx | (xx << 2)) & 0x3333333333333333
	xx = (xx | (xx << 1)) & 0x5555555555555555

	// Spread the bits of y into the even positions and shift to odd positions
	yy = (yy | (yy << 16)) & 0x0000FFFF0000FFFF
	yy = (yy | (yy << 8)) & 0x00FF00FF00FF00FF
	yy = (yy | (yy << 4)) & 0x0F0F0F0F0F0F0F0F
	yy = (yy | (yy << 2)) & 0x3333333333333333
	yy = (yy | (yy << 1)) & 0x5555555555555555

	// Combine x (even bits) and y (odd bits)
	return xx | (yy << 1)
}

type BvhTreeComponentManager = ecs.ComponentManager[BvhTree]

func NewBvhTreeComponentManager() BvhTreeComponentManager {
	return ecs.NewComponentManager[BvhTree](BvhTreeComponentId)
}
