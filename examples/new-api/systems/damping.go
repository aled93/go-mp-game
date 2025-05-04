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

const (
	dampingFactor float32 = 0.98
)

func (s *DampingSystem) Init() {}

func (s *DampingSystem) Run(dt time.Duration) {
	s.Velocities.EachEntity()(func(e ecs.Entity) bool {
		velocity := s.Velocities.GetUnsafe(e)
		rigidbody := s.RigidBodies.GetUnsafe(e)

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
