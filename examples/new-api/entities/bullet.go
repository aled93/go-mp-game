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
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/pkg/util"
	"gomp/stdcomponents"
	"image/color"
)

type CreateBulletManagers struct {
	EntityManager   *ecs.EntityManager
	Positions       *stdcomponents.PositionComponentManager
	Rotations       *stdcomponents.RotationComponentManager
	Scales          *stdcomponents.ScaleComponentManager
	Velocities      *stdcomponents.VelocityComponentManager
	CircleColliders *stdcomponents.CircleColliderComponentManager
	BoxColliders    *stdcomponents.BoxColliderComponentManager
	RigidBodies     *stdcomponents.RigidBodyComponentManager
	Sprites         *stdcomponents.SpriteComponentManager
	BulletTags      *components.BulletTagComponentManager
	Hps             *components.HpComponentManager
	Renderables     *stdcomponents.RenderableComponentManager
}

func CreateBullet(
	props CreateBulletManagers,
	posX, posY float32,
	angle float64,
	velocityX, velocityY float32,
) ecs.Entity {
	entity := props.EntityManager.Create()
	props.Positions.Create(entity, stdcomponents.Position{
		XY: util.NewVec2(posX, posY),
	})
	props.Rotations.Create(entity, stdcomponents.Rotation{}.SetFromDegrees(angle))
	props.Scales.Create(entity, stdcomponents.Scale{
		XY: util.NewVec2(1, 1),
	})
	props.Velocities.Create(entity, stdcomponents.Velocity{
		X: velocityX,
		Y: velocityY,
	})
	props.CircleColliders.Create(entity, stdcomponents.CircleCollider{
		Radius:     6,
		Offset:     util.NewVec2(0, 0),
		Layer:      config.BulletCollisionLayer,
		Mask:       1<<config.EnemyCollisionLayer | 1<<config.WallCollisionLayer | 1<<config.BulletCollisionLayer,
		AllowSleep: true,
	})
	//props.BoxColliders.Create(entity, stdcomponents.BoxCollider{
	//	WH: util.Vec2{
	//		X: 16,
	//		Y: 16,
	//	},
	//	Offset: util.Vec2{
	//		X: 8,
	//		Y: 8,
	//	},
	//	Layer:      config.BulletCollisionLayer,
	//	Mask:       1<<config.EnemyCollisionLayer | 1<<config.WallCollisionLayer | 1<<config.BulletCollisionLayer,
	//	AllowSleep: true,
	//})
	te := stdcomponents.Sprite{
		Texture: assets.Textures.Get("bullet.png"),
		Frame:   util.NewRectFromOriginSize(util.NewVec2(0, 0), util.NewVec2(64, 64)),
		Origin:  util.NewVec2(32, 32),
		Tint: color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		},
	}
	props.Sprites.Create(entity, te)
	props.BulletTags.Create(entity, components.BulletTag{})
	props.Hps.Create(entity, components.Hp{
		Hp:    1,
		MaxHp: 1,
	})
	props.RigidBodies.Create(entity, stdcomponents.RigidBody{
		IsStatic: false,
		Mass:     1,
	})
	props.Renderables.Create(entity, stdcomponents.Renderable{
		Type:       stdcomponents.SpriteRenderableType,
		CameraMask: config.MainCameraLayer,
	})

	return entity
}
