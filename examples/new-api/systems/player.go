/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/examples/new-api/entities"
	"gomp/examples/new-api/sprites"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math/rand"
)

func NewPlayerSystem() PlayerSystem {
	return PlayerSystem{}
}

type PlayerSystem struct {
	EntityManager    *ecs.EntityManager
	SpriteMatrixes   *stdcomponents.SpriteMatrixComponentManager
	Positions        *stdcomponents.PositionComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	Velocities       *stdcomponents.VelocityComponentManager
	AnimationPlayers *stdcomponents.AnimationPlayerComponentManager
	AnimationStates  *stdcomponents.AnimationStateComponentManager
	Tints            *stdcomponents.TintComponentManager
	Flips            *stdcomponents.FlipComponentManager
	HP               *components.HealthComponentManager
	Controllers      *components.ControllerComponentManager
	Renderables      *stdcomponents.RenderableComponentManager
	YSorts           *stdcomponents.YSortComponentManager
	RenderOrders     *stdcomponents.RenderOrderComponentManager
	BoxColliders     *stdcomponents.ColliderBoxComponentManager
	GenericCollider  *stdcomponents.GenericColliderComponentManager
	Players          *components.PlayerTagComponentManager
}

func (s *PlayerSystem) Init() {
	s.SpriteMatrixes.Create(sprites.PlayerSpriteSharedComponentId, sprites.PlayerSpriteMatrix)

	for range 5_000 {
		npc := entities.CreatePlayer(
			s.EntityManager, s.SpriteMatrixes, s.Positions, s.Rotations, s.Scales,
			s.Velocities, s.AnimationPlayers, s.AnimationStates, s.Tints, s.Flips, s.Renderables,
			s.YSorts, s.RenderOrders, s.BoxColliders, s.GenericCollider,
		)

		npc.Position.X = 100 + rand.Float32()*700
		npc.Position.Y = 100 + rand.Float32()*500
		npc.AnimationPlayer.Current = uint8(rand.Intn(7))
	}

	player := entities.CreatePlayer(
		s.EntityManager, s.SpriteMatrixes, s.Positions, s.Rotations, s.Scales,
		s.Velocities, s.AnimationPlayers, s.AnimationStates, s.Tints, s.Flips, s.Renderables,
		s.YSorts, s.RenderOrders, s.BoxColliders, s.GenericCollider,
	)
	player.Position.X = 100
	player.Position.Y = 100
	player.GenericCollider.Layer = config.PlayerCollisionLayer
	player.GenericCollider.Mask = 1 << config.EnemyCollisionLayer

	s.Controllers.Create(player.Entity, components.Controller{})
	s.Players.Create(player.Entity, components.PlayerTag{})

}
func (s *PlayerSystem) Run() {

	var speed float32 = 300

	for e := range s.Controllers.EachEntity {
		velocity := s.Velocities.Get(e)
		flip := s.Flips.Get(e)
		animationState := s.AnimationStates.Get(e)

		velocity.X = 0
		velocity.Y = 0

		if rl.IsKeyDown(rl.KeySpace) {
			*animationState = entities.PlayerStateJump
		} else {
			*animationState = entities.PlayerStateIdle
			if rl.IsKeyDown(rl.KeyD) {
				*animationState = entities.PlayerStateWalk
				velocity.X = speed
				flip.X = false
			}
			if rl.IsKeyDown(rl.KeyA) {
				*animationState = entities.PlayerStateWalk
				velocity.X = -speed
				flip.X = true
			}
			if rl.IsKeyDown(rl.KeyW) {
				*animationState = entities.PlayerStateWalk
				velocity.Y = -speed
			}
			if rl.IsKeyDown(rl.KeyS) {
				*animationState = entities.PlayerStateWalk
				velocity.Y = speed
			}
		}

		if rl.IsKeyPressed(rl.KeyK) {
			s.EntityManager.Delete(e)
		}
	}

}
func (s *PlayerSystem) Destroy() {}
