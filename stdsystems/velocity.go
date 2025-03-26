/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewVelocitySystem() VelocitySystem {
	return VelocitySystem{}
}

type VelocitySystem struct {
	Velocities  *stdcomponents.VelocityComponentManager
	Positions   *stdcomponents.PositionComponentManager
	RigidBodies *stdcomponents.RigidBodyComponentManager
}

func (s *VelocitySystem) Init() {}

func (s *VelocitySystem) Run(dt time.Duration) {
	dtSec := float32(dt.Seconds())
	dampingFactor := float32(0.98) // Damping factor for velocity

	s.Velocities.EachEntity(func(e ecs.Entity) bool {
		velocity := s.Velocities.Get(e)
		position := s.Positions.Get(e)
		rigidbody := s.RigidBodies.Get(e)

		position.XY.X += velocity.X * dtSec
		position.XY.Y += velocity.Y * dtSec

		if rigidbody != nil && !rigidbody.IsStatic {
			velocity.X *= dampingFactor
			velocity.Y *= dampingFactor
		}
		return true
	})
}

func (s *VelocitySystem) Destroy() {}
