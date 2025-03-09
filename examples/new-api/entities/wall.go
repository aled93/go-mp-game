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
)

type CreateWallManagers struct {
	EntityManager *ecs.EntityManager
	Positions     *stdcomponents.PositionComponentManager
	Rotations     *stdcomponents.RotationComponentManager
	Scales        *stdcomponents.ScaleComponentManager
	BoxColliders  *stdcomponents.BoxColliderComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	WallTags      *components.WallTagComponentManager
}

func CreateWall(
	props CreateWallManagers,
	posX, posY, angle float32,
	width, height float32,
) ecs.Entity {
	entity := props.EntityManager.Create()
	props.Positions.Create(entity, stdcomponents.Position{
		X: posX,
		Y: posY,
		Z: 0,
	})
	props.Rotations.Create(entity, stdcomponents.Rotation{
		Angle: angle,
	})
	props.Scales.Create(entity, stdcomponents.Scale{
		X: 1,
		Y: 1,
	})
	props.BoxColliders.Create(entity, stdcomponents.BoxCollider{
		Width:   width,
		Height:  height,
		OffsetX: 0,
		OffsetY: 0,
		Layer:   config.WallCollisionLayer,
		Mask:    1 << config.PlayerCollisionLayer,
	})
	props.Sprites.Create(entity, stdcomponents.Sprite{
		Texture: assets.Textures.Get("wall.png"),
		Frame: rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
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
	props.WallTags.Create(entity, components.Wall{})

	return entity
}
