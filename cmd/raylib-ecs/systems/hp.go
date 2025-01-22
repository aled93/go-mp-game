/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
)

type hpController struct {
	healths  *ecs.ComponentManager[components.Health]
	spatials *ecs.ComponentManager[components.SpatialEnt]
}

func (s *hpController) Init(world *ecs.World) {
	s.healths = components.HealthService.GetManager(world)
	s.spatials = components.SpatialEntService.GetManager(world)
}

func (s *hpController) Update(world *ecs.World) {}
func (s *hpController) FixedUpdate(world *ecs.World) {
	s.healths.All(func(entity ecs.EntityID, h *components.Health) bool {
		h.Hp--

		if h.Hp <= 0 {
			if spatial := s.spatials.Get(entity); spatial != nil {
				gravity.QTree.Remove(spatial.Ent)
			}
			world.DestroyEntity(entity)
		}

		return true
	})
}
func (s *hpController) Destroy(world *ecs.World) {}
