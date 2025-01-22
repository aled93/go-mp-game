/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/pkgs/gomp/ecs"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type screenWrapController struct {
	positions *ecs.ComponentManager[components.Position]
}

func (s *screenWrapController) Init(world *ecs.World) {
	s.positions = components.PositionService.GetManager(world)
}

func (s *screenWrapController) Update(world *ecs.World) {
	w, h := float32(rl.GetFramebufferWidth()), float32(rl.GetFramebufferHeight())

	s.positions.AllParallel(func(entId ecs.EntityID, pos *components.Position) bool {
		if pos.X < 0.0 {
			pos.X = w
		} else if pos.X > w {
			pos.X = 0.0
		}

		if pos.Y < 0.0 {
			pos.Y = h
		} else if pos.Y > h {
			pos.Y = 0.0
		}

		return true
	})
}

func (s *screenWrapController) FixedUpdate(world *ecs.World) {}
func (s *screenWrapController) Destroy(world *ecs.World)     {}
