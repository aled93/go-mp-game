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

type spritePosController struct {
	sprites    ecs.WorldComponents[components.Sprite]
	transforms ecs.WorldComponents[components.Transform]
}

func (s *spritePosController) Init(world *ecs.World) {
	s.sprites = components.SpriteService.GetManager(world)
	s.transforms = components.TransformService.GetManager(world)
}

func (s *spritePosController) Update(world *ecs.World) {
	s.sprites.AllParallel(func(ei ecs.EntityID, spr *components.Sprite) bool {
		t := s.transforms.GetPtr(ei)
		if t == nil {
			return true
		}

		spr.Pos.X = t.X
		spr.Pos.Y = t.Y

		return true
	})
}

func (s *spritePosController) FixedUpdate(world *ecs.World) {}
func (s *spritePosController) Destroy(world *ecs.World)     {}
