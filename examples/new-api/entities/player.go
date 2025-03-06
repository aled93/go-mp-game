/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package entities

import (
	"gomp/examples/new-api/sprites"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

const (
	PlayerStateIdle stdcomponents.AnimationState = iota
	PlayerStateWalk
	PlayerStateJump
	PlayerStateFall
	PlayerStateAttack
	PlayerStateHurt
	PlayerStateDie
)

type Player struct {
	ecs.Entity
	Position        *stdcomponents.Position
	Rotation        *stdcomponents.Rotation
	Scale           *stdcomponents.Scale
	Velocity        *stdcomponents.Velocity
	SpriteMatrix    *stdcomponents.SpriteMatrix
	Tint            *stdcomponents.Tint
	AnimationPlayer *stdcomponents.AnimationPlayer
	AnimationState  *stdcomponents.AnimationState
	Flip            *stdcomponents.Flip
	Renderable      *stdcomponents.Renderable
	YSort           *stdcomponents.YSort
	RenderOrder     *stdcomponents.RenderOrder
	ColliderBox     *stdcomponents.ColliderBox
	GenericCollider *stdcomponents.GenericCollider
}

func CreatePlayer(
	world *ecs.EntityManager,
	spriteMatrixes *stdcomponents.SpriteMatrixComponentManager,
	positions *stdcomponents.PositionComponentManager,
	rotations *stdcomponents.RotationComponentManager,
	scales *stdcomponents.ScaleComponentManager,
	velocities *stdcomponents.VelocityComponentManager,
	animationPlayers *stdcomponents.AnimationPlayerComponentManager,
	animationStates *stdcomponents.AnimationStateComponentManager,
	tints *stdcomponents.TintComponentManager,
	flips *stdcomponents.FlipComponentManager,
	renderables *stdcomponents.RenderableComponentManager,
	ySorts *stdcomponents.YSortComponentManager,
	renderOrders *stdcomponents.RenderOrderComponentManager,
	boxColliders *stdcomponents.ColliderBoxComponentManager,
	genericColliders *stdcomponents.GenericColliderComponentManager,
) (player Player) {
	// Creating new player

	entity := world.Create()
	player.Entity = entity

	// Adding position component
	t := stdcomponents.Position{}
	player.Position = positions.Create(entity, t)

	// Adding rotation component
	rotation := stdcomponents.Rotation{}
	player.Rotation = rotations.Create(entity, rotation)

	// Adding scale component
	scale := stdcomponents.Scale{
		X: 1,
		Y: 1,
	}
	player.Scale = scales.Create(entity, scale)

	// Adding velocity component
	velocity := stdcomponents.Velocity{}
	player.Velocity = velocities.Create(entity, velocity)

	// Adding Tint component
	tint := stdcomponents.Tint{R: 255, G: 255, B: 255, A: 255}
	player.Tint = tints.Create(entity, tint)

	// Adding sprite matrix component
	player.SpriteMatrix = spriteMatrixes.Set(entity, sprites.PlayerSpriteSharedComponentId)

	// Adding animation player component
	animation := stdcomponents.AnimationPlayer{}
	player.AnimationPlayer = animationPlayers.Create(entity, animation)

	// Adding Animation state component
	player.AnimationState = animationStates.Create(entity, PlayerStateWalk)

	// Adding Flip component
	player.Flip = flips.Create(entity, stdcomponents.Flip{})

	// Adding renderable component
	player.Renderable = renderables.Create(entity, stdcomponents.SpriteMatrixRenderableType)

	// Adding YSort component
	player.YSort = ySorts.Create(entity, stdcomponents.YSort{})

	// Adding RenderOrder component
	player.RenderOrder = renderOrders.Create(entity, stdcomponents.RenderOrder{})

	// Adding ColliderBox component
	player.ColliderBox = boxColliders.Create(entity, stdcomponents.ColliderBox{
		Width:  96,
		Height: 128,
	})

	// Adding GenericCollider component
	player.GenericCollider = genericColliders.Create(entity, stdcomponents.GenericCollider{
		Shape: stdcomponents.BoxColliderShape,
	})

	return player
}
