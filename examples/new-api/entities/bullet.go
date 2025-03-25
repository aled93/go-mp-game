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
	"gomp/vectors"
	"image/color"
)

type CreateBulletManagers struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	Rotations       *stdcomponents.RotationComponentManager
	Scales          *stdcomponents.ScaleComponentManager
	Velocities      *stdcomponents.VelocityComponentManager
	CircleColliders *stdcomponents.CircleColliderComponentManager
	Sprites         *stdcomponents.SpriteComponentManager
	BulletTags      *components.BulletTagComponentManager
	Hps             *components.HpComponentManager
}

func CreateBullet(
	props CreateBulletManagers,
	posX, posY float32,
	angle float64,
	velocityX, velocityY float32,
) ecs.Entity {
	bullet := props.EntityManager.Create()
	props.Positions.Create(bullet, stdcomponents.Position{
		XY: vectors.Vec2{
			X: posX,
			Y: posY,
		},
	})
	props.Rotations.Create(bullet, stdcomponents.Rotation{}.SetFromDegrees(angle))
	props.Scales.Create(bullet, stdcomponents.Scale{
		XY: vectors.Vec2{
			X: 1,
			Y: 1,
		},
	})
	props.Velocities.Create(bullet, stdcomponents.Velocity{
		X: velocityX,
		Y: velocityY,
	})
	props.CircleColliders.Create(bullet, stdcomponents.CircleCollider{
		Radius: 8,
		Offset: vectors.Vec2{
			X: 0,
			Y: 0,
		},
		Layer: config.BulletCollisionLayer,
		Mask:  1<<config.EnemyCollisionLayer | 1<<config.WallCollisionLayer,
	})
	props.Sprites.Create(bullet, stdcomponents.Sprite{
		Texture: assets.Textures.Get("bullet.png"),
		Frame: rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  64,
			Height: 64,
		},
		Origin: rl.Vector2{
			X: 32,
			Y: 32,
		},
		Tint: color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		},
	})
	props.BulletTags.Create(bullet, components.BulletTag{})
	props.Hps.Create(bullet, components.Hp{
		Hp:    1,
		MaxHp: 1,
	})

	return bullet
}
