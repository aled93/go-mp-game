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
	var moveSpeedMax float32 = 300
	var moveSpeedMaxBackwards float32 = -200
	var rotateSpeed float32 = 10
	var speedIncrement float32 = 10

	var bulletSpeed float32 = 300

	s.SpaceshipIntents.EachEntity(func(entity ecs.Entity) bool {
		intent := s.SpaceshipIntents.Get(entity)
		vel := s.Velocities.Get(entity)
		rot := s.Rotations.Get(entity)
		pos := s.Positions.Get(entity)
		weapon := s.Weapons.Get(entity)

		//if pos.Y < 0 || pos.Y > 5000 || pos.X < 0 || pos.X > 5000 {
		//	vel.X *= -1
		//	vel.Y *= -1
		//	rot.Angle += 180
		//}

		if intent.RotateLeft {
			rot.Angle -= rotateSpeed
		}
		if intent.RotateRight {
			rot.Angle += rotateSpeed
		}
		if intent.MoveUp {
			s.moveSpeed += speedIncrement
			if s.moveSpeed > moveSpeedMax {
				s.moveSpeed = moveSpeedMax
			}
		}
		if intent.MoveDown {
			s.moveSpeed -= speedIncrement
			if s.moveSpeed < moveSpeedMaxBackwards {
				s.moveSpeed = moveSpeedMaxBackwards
			}
		}

		if !intent.MoveUp && !intent.MoveDown {
			if s.moveSpeed > 0 {
				s.moveSpeed -= speedIncrement
			} else if s.moveSpeed < 0 {
				s.moveSpeed += speedIncrement
			}
		}

		rads := deg2rad(float64(rot.Angle)) + math.Pi

		vel.Y = float32(math.Cos(rads)) * s.moveSpeed
		vel.X = -float32(math.Sin(rads)) * s.moveSpeed

		if weapon.CooldownLeft <= 0 {
			if intent.Fire {
				count := 360
				ofs := 1
				startAngle := rot.Angle - float32(count*ofs/2)
				for i := range count {
					angle := startAngle + float32(i*ofs)
					rads = deg2rad(float64(angle)) + math.Pi

					bulletVelocityY := vel.Y + float32(math.Cos(rads))*bulletSpeed
					bulletVelocityX := vel.X - float32(math.Sin(rads))*bulletSpeed
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
					}, pos.X, pos.Y, angle, bulletVelocityX, bulletVelocityY)
				}
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
