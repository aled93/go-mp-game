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

const uintShift = 7 - 64/bits.UintSize

// nextPowerOf2 rounds up to the next power of 2.
// For example: 5 -> 8, 17 -> 32, 32 -> 32
func nextPowerOf2(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}

func NewComponentBitTable(maxComponentsLen int) ComponentBitTable {
	bitsetSize := ((maxComponentsLen - 1) / bits.UintSize) + 1 + 1 // 1 entry for the entity
	return ComponentBitTable{
		bits:       make([][]uint, 0, initialBookSize),
		lookup:     NewPagedMap[Entity, int](),
		bitsetSize: bitsetSize,
		pageSize:   bitsetSize * pageSize,
	}
}

type ComponentBitTable struct {
	bits       [][]uint
	lookup     PagedMap[Entity, int]
	length     int
	bitsetSize int
	pageSize   int
}

func (b *ComponentBitTable) Create(entity Entity) {
	bitsId, ok := b.lookup.Get(entity)
	assert.False(ok, "entity already exists")
	b.extend()
	bitsId = b.length
	b.lookup.Set(entity, bitsId)
	chunkId := bitsId / b.pageSize
	bitsetId := bitsId % b.pageSize
	b.bits[chunkId][bitsetId] = uint(entity)
	b.length += b.bitsetSize
}

func (b *ComponentBitTable) Delete(entity Entity) {
	bitsId, ok := b.lookup.Get(entity)
	assert.True(ok, "entity not found")

	// Get the index of the last entity
	lastIndex := b.length - b.bitsetSize

	// If this is not the last entity, swap with the last one
	if bitsId != lastIndex {
		lastChunkId := lastIndex / b.pageSize
		lastBitsetId := lastIndex % b.pageSize
		deleteChunkId := bitsId / b.pageSize
		deleteBitsetId := bitsId % b.pageSize

		lastEntity := b.bits[lastChunkId][lastBitsetId]
		for i := 0; i < b.bitsetSize; i++ {
			b.bits[deleteChunkId][deleteBitsetId+i] = b.bits[lastChunkId][lastBitsetId+i]
			b.bits[lastChunkId][lastBitsetId+i] = 0
		}

		b.lookup.Set(Entity(lastEntity), bitsId)
	}

	b.lookup.Delete(entity)
	b.length -= b.bitsetSize
}

// Set sets the bit at the given index to 1.
func (b *ComponentBitTable) Set(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup.Get(entity)
	assert.True(ok, "entity not found")

	chunkId := bitsId / b.pageSize
	bitsetId := bitsId % b.pageSize
	offset := int(componentId>>uintShift) + 1 // +1 to skip the first Entity entry
	b.bits[chunkId][bitsetId+offset] |= 1 << (componentId % bits.UintSize)
}

// Unset clears the bit at the given index (sets it to 0).
func (b *ComponentBitTable) Unset(entity Entity, componentId ComponentId) {
	bitsId, ok := b.lookup.Get(entity)
	assert.True(ok, "entity not found")
	chunkId := bitsId / b.pageSize
	bitsetId := bitsId % b.pageSize
	offset := int(componentId>>uintShift) + 1 // +1 to skip the first Entity entry
	b.bits[chunkId][bitsetId+offset] &= ^(1 << (componentId % bits.UintSize))
}

func (b *ComponentBitTable) Test(entity Entity, componentId ComponentId) bool {
	bitsId, ok := b.lookup.Get(entity)
	if !ok {
		return false
	}
	chunkId := bitsId / b.pageSize
	bitsetId := bitsId % b.pageSize
	offset := int(componentId>>uintShift) + 1 // +1 to skip the first Entity entry
	return (b.bits[chunkId][bitsetId+offset] & (1 << (componentId % bits.UintSize))) != 0
}

func (b *ComponentBitTable) AllSet(entity Entity, yield func(ComponentId) bool) {
	bitsId, ok := b.lookup.Get(entity)
	if !ok {
		return
	}
	chunkId := bitsId / b.pageSize
	bitsetId := bitsId % b.pageSize
	for i := 1; i < b.bitsetSize; i++ { // i := 1 Skip the first entry (Entity)
		for j := 0; j < bits.UintSize; j++ {
			if (b.bits[chunkId][bitsetId+i]>>j)&1 == 1 {
				if !yield(ComponentId((i-1)*bits.UintSize + j)) {
					return
				}
			}
		}
	}
}

func (b *ComponentBitTable) extend() {
	lastChunkId := b.length / b.pageSize
	if lastChunkId == len(b.bits) && b.length%b.pageSize == 0 {
		b.bits = append(b.bits, make([]uint, b.pageSize))
	}
}
