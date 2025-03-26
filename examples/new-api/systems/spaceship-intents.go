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
	"gomp/vectors"
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
	CircleColliders  *stdcomponents.CircleColliderComponentManager
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
	var rotateSpeed vectors.Radians = 3
	var speedIncrement float32 = 10

	var bulletSpeed float32 = 300

	dtSec := float32(dt.Seconds())

	s.SpaceshipIntents.EachEntity(func(entity ecs.Entity) bool {
		intent := s.SpaceshipIntents.Get(entity)
		vel := s.Velocities.Get(entity)
		rot := s.Rotations.Get(entity)
		pos := s.Positions.Get(entity)
		weapon := s.Weapons.Get(entity)

		if intent.RotateLeft {
			rot.Angle -= rotateSpeed * vectors.Radians(dtSec)
		}
		if intent.RotateRight {
			rot.Angle += rotateSpeed * vectors.Radians(dtSec)
		}
		if intent.MoveUp {
			vel.Y += float32(math.Cos(rot.Angle+math.Pi)) * speedIncrement
			vel.X -= float32(math.Sin(rot.Angle+math.Pi)) * speedIncrement
			if vel.Vec2().Length() > moveSpeedMax {
				vel.SetVec2(vel.Vec2().Normalize().Scale(moveSpeedMax))
			}
		}
		if intent.MoveDown {
			vel.Y -= float32(math.Cos(rot.Angle+math.Pi)) * speedIncrement
			vel.X += float32(math.Sin(rot.Angle+math.Pi)) * speedIncrement
			if vel.Vec2().Length() < moveSpeedMaxBackwards {
				vel.SetVec2(vel.Vec2().Normalize().Scale(moveSpeedMaxBackwards))
			}
		}

		if !intent.MoveUp && !intent.MoveDown {
			if vel.Vec2().Length() > 0 {
				deceleration := vel.Vec2().Normalize().Scale(speedIncrement)
				vel.SetVec2(vel.Vec2().Sub(deceleration))
				if vel.Vec2().Length() < speedIncrement {
					vel.SetVec2(vectors.Vec2{0, 0})
				}
			}
		}

		if weapon.CooldownLeft <= 0 {
			if intent.Fire {
				var count int = 360
				for i := 0; i < count; i++ {
					var angle = math.Pi*2/float64(count)*float64(i) + rot.Angle - math.Pi/2

					bulletVelocityY := vel.Y + float32(math.Cos(angle+math.Pi))*bulletSpeed
					bulletVelocityX := vel.X - float32(math.Sin(angle+math.Pi))*bulletSpeed
					entities.CreateBullet(entities.CreateBulletManagers{
						EntityManager:   s.EntityManager,
						Positions:       s.Positions,
						Rotations:       s.Rotations,
						Scales:          s.Scales,
						Velocities:      s.Velocities,
						CircleColliders: s.CircleColliders,
						Sprites:         s.Sprites,
						BulletTags:      s.BulletTags,
						Hps:             s.Hps,
					}, pos.XY.X, pos.XY.Y, angle, bulletVelocityX, bulletVelocityY)
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
