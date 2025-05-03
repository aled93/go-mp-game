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
	"gomp/pkg/ecs"
	"gomp/pkg/worker"
	"sync"

	"github.com/negrel/assert"
)

const (
	MembersPerCellSqrt = 2
	membersPerCell     = MembersPerCellSqrt * MembersPerCellSqrt
)

func NewMemberListPool(workerPool *worker.Pool) MemberListPool {
	return MemberListPool{
		pool: sync.Pool{
			New: func() any {
				return &MemberList{
					Members:  make([]ecs.Entity, 0, membersPerCell),
					Lookup:   ecs.NewGenMap[ecs.Entity, int](membersPerCell),
					InputAcc: make([][]ecs.Entity, workerPool.NumWorkers()),
				}
			},
		},
	}
}

type MemberListPool struct {
	pool sync.Pool
}

func (p *MemberListPool) Get() *MemberList {
	return p.pool.Get().(*MemberList)
}

func (p *MemberListPool) Put(ml *MemberList) {
	p.pool.Put(ml)
}

type MemberList struct {
	Members  []ecs.Entity
	Lookup   ecs.GenMap[ecs.Entity, int]
	InputAcc [][]ecs.Entity
}

func (ml *MemberList) Add(member ecs.Entity) {
	ml.Members = append(ml.Members, member)
	ml.Lookup.Set(member, len(ml.Members)-1)
}

func (ml *MemberList) Delete(member ecs.Entity) {
	index, ok := ml.Lookup.Get(member)
	assert.True(ok)
	lastIndex := len(ml.Members) - 1
	if index < lastIndex {
		// Swap the dead element with the last one
		ml.Members[index], ml.Members[lastIndex] = ml.Members[lastIndex], ml.Members[index]
		// Update Lookup table
		ml.Lookup.Set(ml.Members[index], index)
	}
	ml.Members = ml.Members[:lastIndex]
	ml.Lookup.Delete(member)
}

func (ml *MemberList) Reset() {
	ml.Lookup.Reset()
	ml.Members = ml.Members[:0]
}

func (ml *MemberList) Has(member ecs.Entity) bool {
	return ml.Lookup.Has(member)
}

type CollisionCell struct {
	Index   SpatialCellIndex
	Layer   CollisionLayer
	Grid    ecs.Entity
	Size    float32
	Members *MemberList
}

type CollisionCellComponentManager = ecs.ComponentManager[CollisionCell]

func NewCollisionCellComponentManager() CollisionCellComponentManager {
	return ecs.NewComponentManager[CollisionCell](CollisionCellComponentId)
}
