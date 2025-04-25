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

package ecs

import (
	"github.com/negrel/assert"
	"math/bits"
)

const ComponentBitTablePreallocate = 1024

func NewComponentBitTable(maxComponentsLen int) ComponentBitTable {
	bitsetSize := (maxComponentsLen / bits.UintSize) + 1
	return ComponentBitTable{
		bits:       make([]uint, bitsetSize, bitsetSize*ComponentBitTablePreallocate),
		lookup:     make(map[Entity]int, ComponentBitTablePreallocate),
		bitsetSize: bitsetSize,
	}
}

type ComponentBitTable struct {
	bits       []uint
	lookup     map[Entity]int
	bitsetSize int
}

// Set sets the bit at the given index to 1.
func (b *ComponentBitTable) Set(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup[entity]
	if !ok { // Most likely the entity is new
		b.extend()
		bitsId = len(b.bits)
		b.lookup[entity] = bitsId
		b.bits = append(b.bits, make([]uint, b.bitsetSize)...)
	}

	offset := int(componentId / bits.UintSize)
	b.bits[bitsId+offset] |= 1 << (componentId % bits.UintSize)
}

// Unset clears the bit at the given index (sets it to 0).
func (b *ComponentBitTable) Unset(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup[entity]
	assert.True(ok, "entity not found")
	offset := int(componentId / bits.UintSize)
	b.bits[bitsId+offset] &= ^(1 << (componentId % bits.UintSize))
}

func (b *ComponentBitTable) extend() {
	if len(b.bits) == cap(b.bits) {
		b.bits = append(b.bits, make([]uint, cap(b.bits)*2)...)
	}
}

func (b *ComponentBitTable) AllSet(entity Entity, yield func(ComponentId) bool) {
	bitsId, ok := b.lookup[entity]
	if !ok {
		return
	}
	for i := 0; i < b.bitsetSize; i++ {
		for j := 0; j < bits.UintSize; j++ {
			if (b.bits[bitsId+i]>>j)&1 == 1 {
				if !yield(ComponentId(i*bits.UintSize + j)) {
					return
				}
			}
		}
	}
}
