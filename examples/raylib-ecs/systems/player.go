/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/examples/raylib-ecs/components"
	"gomp/examples/raylib-ecs/entities"
	"gomp/pkg/ecs"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const min_move_walk_anim = 0.1

type playerController struct {
	player   entities.Player
	selector ecs.Selector[struct {
		Position       *components.Position
		InputIntent    *components.InputIntent
		AnimationState *components.AnimationState
		Mirrored       *components.Mirrored
	}]
}

func (s *playerController) Init(world *ecs.World) {
	s.player = entities.CreatePlayer(world)
	s.player.Position.X = 100
	s.player.Position.Y = 100
	components.LocalInputService.GetManager(world).Create(s.player.Entity, components.LocalInput{}) // only for first player

	world.RegisterSelector(&s.selector)
}
func (s *playerController) Update(world *ecs.World) {
	if rl.IsKeyPressed(rl.KeySpace) {
		bot := entities.CreatePlayer(world)
		bot.Position.X = 300
		bot.Position.Y = 300
		components.BotRoamerService.GetManager(world).Create(bot.Entity, components.BotRoamer{
			Chilling:      true,
			ChillDuration: 100,
		})
	}

	for e := range s.selector.All() {
		if false /* e.InputIntent.Jump */ {
			*e.AnimationState = entities.PlayerStateJump
		} else {
			*e.AnimationState = entities.PlayerStateIdle

			if rl.Vector2LengthSqr(e.InputIntent.Move) > min_move_walk_anim*min_move_walk_anim {
				*e.AnimationState = entities.PlayerStateWalk

				if e.InputIntent.Move.X < 0.0 {
					e.Mirrored.X = true
				} else {
					e.Mirrored.X = false
				}

				e.Position.X += e.InputIntent.Move.X
				e.Position.Y += e.InputIntent.Move.Y
			}
		}
	}
}
func (s *playerController) FixedUpdate(world *ecs.World) {}
func (s *playerController) Destroy(world *ecs.World)     {}
