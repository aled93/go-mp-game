/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/pkgs/gomp/ecs"
)

type inertiaController struct {
	positions  *ecs.ComponentManager[components.Position]
	velocities *ecs.ComponentManager[components.Velocity]
}

func (s *inertiaController) Init(world *ecs.World) {
	s.positions = components.PositionService.GetManager(world)
	s.velocities = components.VelocityService.GetManager(world)
}

func (s *inertiaController) Update(world *ecs.World) {}

func (s *inertiaController) FixedUpdate(world *ecs.World) {
	s.velocities.All(func(ei ecs.EntityID, vel *components.Velocity) bool {
		t := s.positions.Get(ei)
		if t == nil {
			return true
		}

		t.X += vel.X
		t.Y += vel.Y

		return true
	})
}

func (s *inertiaController) Destroy(world *ecs.World) {}
