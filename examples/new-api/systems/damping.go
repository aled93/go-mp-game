/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewDampingSystem() DampingSystem {
	return DampingSystem{}
}

type DampingSystem struct {
	Velocities  *stdcomponents.VelocityComponentManager
	Positions   *stdcomponents.PositionComponentManager
	RigidBodies *stdcomponents.RigidBodyComponentManager
}

func (s *DampingSystem) Init() {}

func (s *DampingSystem) Run(dt time.Duration) {
	dampingFactor := float32(0.98) // Damping factor for velocity

	s.Velocities.EachEntity(func(e ecs.Entity) bool {
		velocity := s.Velocities.Get(e)
		rigidbody := s.RigidBodies.Get(e)

		if rigidbody != nil && !rigidbody.IsStatic {
			velocity.X *= dampingFactor
			velocity.Y *= dampingFactor
			if velocity.X < 0.1 && velocity.X > -0.1 {
				velocity.X = 0
			}
			if velocity.Y < 0.1 && velocity.Y > -0.1 {
				velocity.Y = 0
			}
		}
		return true
	})
}

func (s *DampingSystem) Destroy() {}
