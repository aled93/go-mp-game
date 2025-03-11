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

package systems

import (
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/entities"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"math"
	"time"
)

func NewSpaceshipIntentsSystem() SpaceshipIntentsSystem {
	return SpaceshipIntentsSystem{}
}

type SpaceshipIntentsSystem struct {
	EntityManager    *ecs.EntityManager
	SpaceshipIntents *components.SpaceshipIntentComponentManager
	Positions        *stdcomponents.PositionComponentManager
	Velocities       *stdcomponents.VelocityComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	BoxColliders     *stdcomponents.BoxColliderComponentManager
	BulletTags       *components.BulletTagComponentManager
	Sprites          *stdcomponents.SpriteComponentManager
	Weapons          *components.WeaponComponentManager
	Hps              *components.HpComponentManager
	moveSpeed        float32
}

func (s *SpaceshipIntentsSystem) Init() {}
func (s *SpaceshipIntentsSystem) Run(dt time.Duration) {
	var thrustMain float32 = 10.0
	var thrustBackwards float32 = -2.0
	var rotateSpeed float32 = 10

	var bulletSpeed float32 = 300

	s.SpaceshipIntents.EachEntity(func(entity ecs.Entity) bool {
		intent := s.SpaceshipIntents.Get(entity)
		vel := s.Velocities.Get(entity)
		rot := s.Rotations.Get(entity)
		pos := s.Positions.Get(entity)
		weapon := s.Weapons.Get(entity)

		rads := deg2rad(float64(rot.Angle)) + math.Pi

		if pos.Y < 0 || pos.Y > 5000 || pos.X < 0 || pos.X > 5000 {
			vel.X *= -1
			vel.Y *= -1
			rot.Angle += 180
		}

		if intent.RotateLeft {
			rot.Angle -= rotateSpeed
		}
		if intent.RotateRight {
			rot.Angle += rotateSpeed
		}
		if intent.MoveUp {
			vel.Y += float32(math.Cos(rads)) * thrustMain
			vel.X += -float32(math.Sin(rads)) * thrustMain
		}
		if intent.MoveDown {
			vel.Y += float32(math.Cos(rads)) * thrustBackwards
			vel.X += -float32(math.Sin(rads)) * thrustBackwards
		}

		bulletVelocityY := vel.Y + float32(math.Cos(rads))*bulletSpeed
		bulletVelocityX := vel.X - float32(math.Sin(rads))*bulletSpeed

		if weapon.CooldownLeft <= 0 {
			if intent.Fire {
				entities.CreateBullet(entities.CreateBulletManagers{
					EntityManager: s.EntityManager,
					Positions:     s.Positions,
					Rotations:     s.Rotations,
					Scales:        s.Scales,
					Velocities:    s.Velocities,
					BoxColliders:  s.BoxColliders,
					Sprites:       s.Sprites,
					BulletTags:    s.BulletTags,
					Hps:           s.Hps,
				}, pos.X, pos.Y, rot.Angle, bulletVelocityX, bulletVelocityY)
				weapon.CooldownLeft = weapon.Cooldown
			}
		} else {
			weapon.CooldownLeft -= dt
		}

		return true
	})
}

func (s *SpaceshipIntentsSystem) Destroy() {}

func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}
