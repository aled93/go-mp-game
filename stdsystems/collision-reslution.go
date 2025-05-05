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
	"gomp/pkg/util"
	"gomp/stdcomponents"
	"time"

	"github.com/negrel/assert"
)

func NewCollisionResolutionSystem() CollisionResolutionSystem {
	return CollisionResolutionSystem{}
}

type CollisionResolutionSystem struct {
	EntityManager *ecs.EntityManager
	Collisions    *stdcomponents.CollisionComponentManager
	Positions     *stdcomponents.PositionComponentManager
	RigidBodies   *stdcomponents.RigidBodyComponentManager
	Velocities    *stdcomponents.VelocityComponentManager
}

func (s *CollisionResolutionSystem) Init() {}
func (s *CollisionResolutionSystem) Run(dt time.Duration) {
	s.Collisions.EachComponent()(func(collision *stdcomponents.Collision) bool {
		if collision.State == stdcomponents.CollisionStateEnter || collision.State == stdcomponents.CollisionStateStay {
			// Resolve penetration
			var displacement util.Vec2

			// Move entities apart
			rigidbody1 := s.RigidBodies.GetUnsafe(collision.E1)
			rigidbody2 := s.RigidBodies.GetUnsafe(collision.E2)

			if rigidbody1 == nil || rigidbody2 == nil {
				return true
			}

			if !rigidbody1.IsStatic && !rigidbody2.IsStatic {
				// both objects are dynamic
				displacement = collision.Normal.ScaleScalar(collision.Depth * 0.5)
			} else {
				// one of the objects is static
				displacement = collision.Normal.ScaleScalar(collision.Depth)
			}

			if !rigidbody1.IsStatic {
				p1 := s.Positions.GetUnsafe(collision.E1)
				p1d := p1.XY.Subtract(displacement)
				p1.XY.X, p1.XY.Y = p1d.X, p1d.Y
			}

			if !rigidbody2.IsStatic {
				p2 := s.Positions.GetUnsafe(collision.E2)
				p2d := p2.XY.Add(displacement)
				p2.XY.X, p2.XY.Y = p2d.X, p2d.Y
			}

			// Apply impulse resolution
			velocity1 := s.Velocities.GetUnsafe(collision.E1)
			assert.NotNil(velocity1)
			velocity2 := s.Velocities.GetUnsafe(collision.E2)
			assert.NotNil(velocity2)

			relativeVelocity := velocity2.Vec2().Subtract(velocity1.Vec2())
			velocityAlongNormal := relativeVelocity.Dot(collision.Normal)

			if velocityAlongNormal > 0 {
				return true
			}

			e := float32(1.0) // Coefficient of restitution (elasticity)
			j := -(1 + e) * velocityAlongNormal
			j /= 1/rigidbody1.Mass + 1/rigidbody2.Mass

			impulse := collision.Normal.ScaleScalar(j)

			if !rigidbody1.IsStatic {
				velocity1.SetVec2(velocity1.Vec2().Subtract(impulse.DivideScalar(rigidbody1.Mass)))
			}

			if !rigidbody2.IsStatic {
				velocity2.SetVec2(velocity2.Vec2().Add(impulse.DivideScalar(rigidbody2.Mass)))
			}
		}
		return true
	})
}
func (s *CollisionResolutionSystem) Destroy() {}
