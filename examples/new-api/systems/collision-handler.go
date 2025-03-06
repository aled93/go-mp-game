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

package systems

import (
	"gomp/examples/new-api/components"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"log"
	"time"
)

func NewCollisionHandlerSystem() CollisionHandlerSystem {
	return CollisionHandlerSystem{}
}

type CollisionHandlerSystem struct {
	EntityManager *ecs.EntityManager
	Collisions    *stdcomponents.CollisionComponentManager
	Players       *components.PlayerTagComponentManager
}

func (s *CollisionHandlerSystem) Init() {}
func (s *CollisionHandlerSystem) Run(dt time.Duration) {
	s.Collisions.EachComponent(func(collision *stdcomponents.Collision) bool {
		if collision.State == stdcomponents.CollisionStateEnter {
			log.Println("Collision entered")
			e1Tag := s.Players.Get(collision.E1)
			e2Tag := s.Players.Get(collision.E2)

			if e1Tag != nil {
				log.Println("PlayerTag 1 collided with player 2")
				s.EntityManager.Delete(collision.E2)
			}
			if e2Tag != nil {
				log.Println("PlayerTag 2 collided with player 1")
				s.EntityManager.Delete(collision.E1)
			}
		}

		if collision.State == stdcomponents.CollisionStateExit {
			log.Println("Collision exited")
		}

		return true
	})
}
func (s *CollisionHandlerSystem) Destroy() {}
