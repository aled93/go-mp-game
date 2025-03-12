package entities

import (
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PickupManagers struct {
	EntityManager *ecs.EntityManager
	Positions     *stdcomponents.PositionComponentManager
	Velocities    *stdcomponents.VelocityComponentManager
	BoxColliders  *stdcomponents.BoxColliderComponentManager
	Sprites       *stdcomponents.SpriteComponentManager
	Pickups       *components.PickupComponentManager
}

func CreatePickup(
	mans PickupManagers,
	posX, posY float32,
	velX, velY float32,
	pickup components.Pickup,
) ecs.Entity {
	ent := mans.EntityManager.Create()
	println("Pickup created")

	mans.Positions.Create(ent, stdcomponents.Position{
		X: posX,
		Y: posY,
	})

	mans.Velocities.Create(ent, stdcomponents.Velocity{
		X: velX,
		Y: velY,
	})

	mans.BoxColliders.Create(ent, stdcomponents.BoxCollider{
		Width:   64,
		Height:  64,
		OffsetX: 32,
		OffsetY: 32,
		Layer:   config.PickupCollisionLayer,
		Mask:    1 << config.PlayerCollisionLayer,
	})

	mans.Sprites.Create(ent, stdcomponents.Sprite{
		Texture: assets.Textures.Get("pickup_" + string(pickup.Power) + ".png"),
		Frame: rl.Rectangle{
			Width:  64,
			Height: 64,
		},
		Origin: rl.Vector2{
			X: 32,
			Y: 32,
		},
	})

	mans.Pickups.Create(ent, pickup)

	return ent
}
