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
	"gomp/vectors"
	"time"
)

func NewCollisionResolutionSystem() CollisionResolutionSystem {
	return CollisionResolutionSystem{}
}

type CollisionResolutionSystem struct {
	EntityManager *ecs.EntityManager
	Collisions    *stdcomponents.CollisionComponentManager
	Positions     *stdcomponents.PositionComponentManager
	RigidBodies   *stdcomponents.RigidBodyComponentManager
}

func (s *CollisionResolutionSystem) Init() {}
func (s *CollisionResolutionSystem) Run(dt time.Duration) {
	s.Collisions.EachComponent(func(collision *stdcomponents.Collision) bool {
		if collision.State == stdcomponents.CollisionStateEnter || collision.State == stdcomponents.CollisionStateStay {
			// Resolve penetration
			var displacement vectors.Vec2

			// Move entities apart
			rigidbody1 := s.RigidBodies.Get(collision.E1)
			rigidbody2 := s.RigidBodies.Get(collision.E2)

			if rigidbody1 == nil || rigidbody2 == nil {
				return true
			}

			if !rigidbody1.IsStatic && !rigidbody2.IsStatic {
				// both objects are dynamic
				displacement = collision.Normal.Scale(collision.Depth * 0.5)
			} else {
				// one of the objects is static
				displacement = collision.Normal.Scale(collision.Depth)
			}

			if !rigidbody1.IsStatic {
				p1 := s.Positions.Get(collision.E1)
				p1d := p1.XY.Sub(displacement)
				p1.XY.X, p1.XY.Y = p1d.X, p1d.Y
			}

			if !rigidbody2.IsStatic {
				p2 := s.Positions.Get(collision.E2)
				p2d := p2.XY.Add(displacement)
				p2.XY.X, p2.XY.Y = p2d.X, p2d.Y
			}

			// Apply forces or velocity changes
		}
		return true
	})
}
func (s *CollisionResolutionSystem) Destroy() {}
