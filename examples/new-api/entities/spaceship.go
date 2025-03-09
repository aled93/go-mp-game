/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package entities

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"image/color"
	"time"
)

type CreateSpaceShipManagers struct {
	EntityManager *ecs.EntityManager
	Positions     *stdcomponents.PositionComponentManager
	Rotations     *stdcomponents.RotationComponentManager
	Scales        *stdcomponents.ScaleComponentManager
	Velocities    *stdcomponents.VelocityComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	BoxColliders  *stdcomponents.BoxColliderComponentManager

	PlayerTags       *components.PlayerTagComponentManager
	Hps              *components.HpComponentManager
	Weapons          *components.WeaponComponentManager
	SpaceshipIntents *components.SpaceshipIntentComponentManager
}

func CreateSpaceShip(
	props CreateSpaceShipManagers,
	posX, posY float32,
	rotationAngle float32,
) ecs.Entity {
	spaceShip := props.EntityManager.Create()
	props.Positions.Create(spaceShip, stdcomponents.Position{
		X: posX,
		Y: posY,
		Z: 0,
	})

	props.Rotations.Create(spaceShip, stdcomponents.Rotation{
		Angle: rotationAngle,
	})

	props.Scales.Create(spaceShip, stdcomponents.Scale{
		X: 1,
		Y: 1,
	})

	props.Velocities.Create(spaceShip, stdcomponents.Velocity{
		X: 0,
		Y: 0,
	})

	props.Sprites.Create(spaceShip, stdcomponents.Sprite{
		Texture: assets.Textures.Get("ship_E.png"),
		Origin:  rl.Vector2{X: 32, Y: 40},
		Frame:   rl.Rectangle{0, 0, 64, 64},
		Tint:    color.RGBA{255, 255, 255, 255},
	})

	props.BoxColliders.Create(spaceShip, stdcomponents.BoxCollider{
		Width:   32,
		Height:  32,
		OffsetX: 16,
		OffsetY: 16,
		Layer:   config.PlayerCollisionLayer,
		Mask:    1<<config.EnemyCollisionLayer | 1<<config.WallCollisionLayer,
	})

	props.PlayerTags.Create(spaceShip, components.PlayerTag{})
	props.Hps.Create(spaceShip, components.Hp{
		Hp:    3,
		MaxHp: 3,
	})

	props.Weapons.Create(spaceShip, components.Weapon{
		Damage:       1,
		Cooldown:     time.Millisecond * 200,
		CooldownLeft: 0,
	})

	props.SpaceshipIntents.Create(spaceShip, components.SpaceshipIntent{})

	return spaceShip
}
