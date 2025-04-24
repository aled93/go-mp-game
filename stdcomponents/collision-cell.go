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
)

type CollisionCell struct {
	//Members      ecs.PagedArray[ecs.Entity]
	//MemberLookup ecs.PagedMap[ecs.Entity, int]

	InputAccumulator []ecs.PagedArray[ecs.Entity]
	Size             float32
	Layer            CollisionLayer
}

func (c *CollisionCell) Init(size float32, layer CollisionLayer, pool *worker.Pool) {
	//c.Members = ecs.NewPagedArray[ecs.Entity]()
	//c.MemberLookup = ecs.NewPagedMap[ecs.Entity, int]()
	c.InputAccumulator = make([]ecs.PagedArray[ecs.Entity], pool.NumWorkers())
	for i := 0; i < pool.NumWorkers(); i++ {
		c.InputAccumulator[i] = ecs.NewPagedArray[ecs.Entity]()
	}
	c.Size = size
	c.Layer = layer
}

//func (c *CollisionCell) AddMember(entity ecs.Entity) {
//	c.Members.Append(entity)
//	c.MemberLookup.Set(entity, c.Members.Len()-1)
//}
//
//func (c *CollisionCell) RemoveMember(entity ecs.Entity) {
//	index, ok := c.MemberLookup.Get(entity)
//	assert.True(ok)
//	c.Members.Swap(index, c.Members.Len()-1)
//	c.Members.SoftReduce()
//	c.MemberLookup.Delete(entity)
//}

type CollisionCellComponentManager = ecs.ComponentManager[CollisionCell]

func NewCollisionCellComponentManager() CollisionCellComponentManager {
	return ecs.NewComponentManager[CollisionCell](CollisionCellComponentId)
}
