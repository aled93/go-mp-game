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

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"sync"
	"time"
)

func NewCollisionSystem() CollisionSystem {
	return CollisionSystem{}
}

type CollisionSystem struct {
	EntityManager *ecs.EntityManager
	Positions     *stdcomponents.PositionComponentManager
	Collisions    *stdcomponents.CollisionComponentManager
	BoxColliders  *stdcomponents.ColliderBoxComponentManager
}

func (s *CollisionSystem) Init() {}
func (s *CollisionSystem) Run(dt time.Duration) {
	var entities []ecs.Entity = make([]ecs.Entity, 0, s.Collisions.Len())
	entities = s.Collisions.RawEntities(entities)

	wg := &sync.WaitGroup{}
	wg.Add(len(entities))

	// Simple AABB collision detection
	for i := 0; i < len(entities); i++ {
		collision1 := s.Collisions.Get(entities[i])

		switch collision1.Shape {
		case stdcomponents.BoxColliderShape:
			go s.boxToXCollision(entities, i, wg)
		case stdcomponents.CircleColliderShape:
			s.circleToXCollision(entities, i)
		default:
			panic("Unknown collision shape")
		}
	}

	wg.Wait()
}
func (s *CollisionSystem) Destroy() {}
func (s *CollisionSystem) boxToXCollision(entities []ecs.Entity, i int, wg *sync.WaitGroup) {
	defer wg.Done()

	position1 := s.Positions.Get(entities[i])
	collider1 := s.BoxColliders.Get(entities[i])

	for j := i + 1; j < len(entities); j++ {
		position2 := s.Positions.Get(entities[j])
		collision2 := s.Collisions.Get(entities[j])
		collider2 := s.BoxColliders.Get(entities[j])

		switch collision2.Shape {
		case stdcomponents.BoxColliderShape:
			// Check AABB collision
			if !(position1.X+collider1.Width < position2.X ||
				position1.X > position2.X+collider2.Width ||
				position1.Y+collider1.Height < position2.Y ||
				position1.Y > position2.Y+collider2.Height) {
				// Handle collision
				//fmt.Printf("Collision between entities %d and %d\n", entities[i], entities[j])
			}
		case stdcomponents.CircleColliderShape:
			panic("Circle-Box collision not implemented")
		default:
			panic("Unknown collision shape")
		}
	}
}

func (s *CollisionSystem) circleToXCollision(entities []ecs.Entity, i int) {
	panic("Circle-X collision not implemented")
}
