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
)

//const (
//	pageSizeShift   = 10
//	pageSize        = 1 << pageSizeShift
//	initialBookSize = 1 // Starting with a small initial book size
//)

func NewComponentByteTable(maxComponentsLen int) ComponentByteTable {
	return ComponentByteTable{
		bytes:          make([][]bool, 0, initialBookSize),
		lookup:         make(map[Entity]int, initialBookSize),
		bytesArraySize: maxComponentsLen,
	}
}

type ComponentByteTable struct {
	bytes          [][]bool
	lookup         map[Entity]int
	length         int
	bytesArraySize int
}

func (b *ComponentByteTable) Create(entity Entity) {
	bytesId, ok := b.lookup[entity]
	if !ok {
		b.extend()
		bytesId = b.length
		b.lookup[entity] = bytesId
		b.length += b.bytesArraySize
	}
}

// Set sets the bit at the given index to 1.
func (b *ComponentByteTable) Set(entity Entity, componentId ComponentId) {
	bytesId, ok := b.lookup[entity]
	if !ok {
		b.extend()
		bytesId = b.length
		b.lookup[entity] = bytesId
		b.length += b.bytesArraySize
	}
	chunkId := bytesId >> pageSizeShift
	index := bytesId % pageSize
	b.bytes[chunkId][index+int(componentId)] = true
}

// Unset clears the bit at the given index (sets it to 0).
func (b *ComponentByteTable) Unset(entity Entity, componentId ComponentId) {
	bytesId, ok := b.lookup[entity]
	assert.True(ok, "entity not found")
	chunkId := bytesId >> pageSizeShift
	index := bytesId % pageSize
	b.bytes[chunkId][index+int(componentId)] = false
}

func (b *ComponentByteTable) Test(entity Entity, componentId ComponentId) bool {
	bytesId, ok := b.lookup[entity]
	if !ok {
		return false
	}
	chunkId := bytesId >> pageSizeShift
	index := bytesId % pageSize
	return b.bytes[chunkId][index+int(componentId)]
}

func (b *ComponentByteTable) extend() {
	lastChunkId := b.length >> pageSizeShift
	if lastChunkId == len(b.bytes) && b.length%pageSize == 0 {
		b.bytes = append(b.bytes, make([]bool, b.bytesArraySize*pageSize))
	}
}

func (b *ComponentByteTable) AllSet(entity Entity, yield func(ComponentId) bool) {
	bytesId, ok := b.lookup[entity]
	if !ok {
		return
	}
	chunkId := bytesId >> pageSizeShift
	index := bytesId % pageSize
	for i := 0; i < b.bytesArraySize; i++ {
		if b.bytes[chunkId][index+i] {
			if !yield(ComponentId(i)) {
				return
			}
		}
	}
}
