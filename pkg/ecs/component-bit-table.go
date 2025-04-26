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

//const (
//	pageSizeShift   = 10
//	pageSize        = 1 << pageSizeShift
//	initialBookSize = 1 // Starting with a small initial book size
//)

func NewComponentBitTable(maxComponentsLen int) ComponentBitTable {
	bitsetSize := ((maxComponentsLen - 1) / bits.UintSize) + 1
	return ComponentBitTable{
		bits:       make([][]uint, 0, initialBookSize),
		lookup:     make(map[Entity]int, pageSize),
		bitsetSize: bitsetSize,
	}
}

type ComponentBitTable struct {
	bits       [][]uint
	lookup     map[Entity]int
	length     int
	bitsetSize int
}

func (b *ComponentBitTable) Create(entity Entity) {
	bitsId, ok := b.lookup[entity]
	if !ok {
		b.extend()
		bitsId = b.length
		b.lookup[entity] = bitsId
		b.length += b.bitsetSize
	}
}

// Set sets the bit at the given index to 1.
func (b *ComponentBitTable) Set(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup[entity]
	if !ok {
		b.extend()
		bitsId = b.length
		b.lookup[entity] = bitsId
		b.length += b.bitsetSize
	}
	chunkId := bitsId >> pageSizeShift
	bitsetId := bitsId % pageSize
	offset := int(componentId / bits.UintSize)
	b.bits[chunkId][bitsetId+offset] |= 1 << (componentId % bits.UintSize)
}

// Unset clears the bit at the given index (sets it to 0).
func (b *ComponentBitTable) Unset(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup[entity]
	assert.True(ok, "entity not found")
	chunkId := bitsId >> pageSizeShift
	bitsetId := bitsId % pageSize
	offset := int(componentId / bits.UintSize)
	b.bits[chunkId][bitsetId+offset] &= ^(1 << (componentId % bits.UintSize))
}

func (b *ComponentBitTable) Test(entity Entity, componentId ComponentId) bool {
	bitsId, ok := b.lookup[entity]
	if !ok {
		return false
	}
	chunkId := bitsId >> pageSizeShift
	bitsetId := bitsId % pageSize
	offset := int(componentId / bits.UintSize)
	return (b.bits[chunkId][bitsetId+offset] & (1 << (componentId % bits.UintSize))) != 0
}

func (b *ComponentBitTable) extend() {
	lastChunkId := b.length >> pageSizeShift
	if lastChunkId == len(b.bits) && b.length%pageSize == 0 {
		b.bits = append(b.bits, make([]uint, b.bitsetSize*pageSize))
	}
}

func (b *ComponentBitTable) AllSet(entity Entity, yield func(ComponentId) bool) {
	bitsId, ok := b.lookup[entity]
	if !ok {
		return
	}
	chunkId := bitsId >> pageSizeShift
	bitsetId := bitsId % pageSize
	for i := 0; i < b.bitsetSize; i++ {
		for j := 0; j < bits.UintSize; j++ {
			if (b.bits[chunkId][bitsetId+i]>>j)&1 == 1 {
				if !yield(ComponentId(i*bits.UintSize + j)) {
					return
				}
			}
		}
	}
}
