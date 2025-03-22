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
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"image/color"
)

type CreateSatelliteManagers struct {
	EntityManager *ecs.EntityManager
	Positions     *stdcomponents.PositionComponentManager
	Rotations     *stdcomponents.RotationComponentManager
	Scales        *stdcomponents.ScaleComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	BoxColliders  *stdcomponents.BoxColliderComponentManager
	RigidBodies   *stdcomponents.RigidBodyComponentManager
}

func CreateSatellite(
	props CreateSatelliteManagers,
	posX, posY float32,
	angle float64,
) ecs.Entity {
	satellite := props.EntityManager.Create()
	props.Positions.Create(satellite, stdcomponents.Position{
		XY: vectors.Vec2{
			X: posX,
			Y: posY,
		},
	})

	props.Rotations.Create(satellite, stdcomponents.Rotation{}.SetFromDegrees(angle))

	props.Scales.Create(satellite, stdcomponents.Scale{
		XY: vectors.Vec2{
			X: 1,
			Y: 1,
		},
	})

	props.Sprites.Create(satellite, stdcomponents.Sprite{
		Texture: assets.Textures.Get("satellite_B.png"),
		Origin:  rl.Vector2{X: 32, Y: 40},
		Frame:   rl.Rectangle{0, 0, 64, 64},
		Tint:    color.RGBA{255, 255, 255, 255},
	})

	props.BoxColliders.Create(satellite, stdcomponents.BoxCollider{
		WH: vectors.Vec2{
			X: 64,
			Y: 64,
		},
		Offset: vectors.Vec2{
			X: 32,
			Y: 32,
		},
		Layer: config.PlayerCollisionLayer,
		Mask:  1<<config.EnemyCollisionLayer | 1<<config.BulletCollisionLayer | 1<<config.PlayerCollisionLayer,
	})
	props.RigidBodies.Create(satellite, stdcomponents.RigidBody{
		IsStatic: false,
	})

	return satellite
}
