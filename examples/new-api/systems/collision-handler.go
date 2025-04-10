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
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/components"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewCollisionHandlerSystem() CollisionHandlerSystem {
	return CollisionHandlerSystem{}
}

type CollisionHandlerSystem struct {
	EntityManager    *ecs.EntityManager
	Collisions       *stdcomponents.CollisionComponentManager
	Players          *components.PlayerTagComponentManager
	PlayerTags       *components.PlayerTagComponentManager
	AsteroidTags     *components.AsteroidComponentManager
	BulletTags       *components.BulletTagComponentManager
	Hps              *components.HpComponentManager
	Weapons          *components.WeaponComponentManager
	SpaceshipIntents *components.SpaceshipIntentComponentManager
	SpaceSpawnerTags *components.SpaceSpawnerComponentManager
	Positions        *stdcomponents.PositionComponentManager
	Rotations        *stdcomponents.RotationComponentManager
	Scales           *stdcomponents.ScaleComponentManager
	Velocities       *stdcomponents.VelocityComponentManager
	Sprites          *stdcomponents.SpriteComponentManager
	BoxColliders     *stdcomponents.BoxColliderComponentManager
	WallTags         *components.WallTagComponentManager
	SoundEffects     *components.SoundEffectsComponentManager
}

func (s *CollisionHandlerSystem) Init() {}
func (s *CollisionHandlerSystem) Run(dt time.Duration) {
	s.Collisions.EachComponent(func(collision *stdcomponents.Collision) bool {
		switch collision.State {
		case stdcomponents.CollisionStateEnter:
			if s.checkBulletCollisionEnter(collision.E1, collision.E2) {
				return true
			}
			if s.checkPlayerCollisionEnter(collision.E1, collision.E2) {
				return true
			}
			if s.checkAsteroidCollisionEnter(collision.E1, collision.E2) {
				return true
			}
		case stdcomponents.CollisionStateExit:
		default:

		}

		return true
	})
}
func (s *CollisionHandlerSystem) Destroy() {}

func (s *CollisionHandlerSystem) checkAsteroidCollisionEnter(e1, e2 ecs.Entity) bool {
	e1Tag := s.AsteroidTags.Get(e1)
	if e1Tag == nil {
		return false
	}

	wallTag := s.WallTags.Get(e2)
	if wallTag != nil {
		hp := s.Hps.Get(e1)
		hp.Hp = 0
		return true
	}

	return false
}

func (s *CollisionHandlerSystem) checkPlayerCollisionEnter(e1, e2 ecs.Entity) bool {
	e1Tag := s.PlayerTags.Get(e1)
	e2Tag := s.PlayerTags.Get(e2)

	if e1Tag != nil {
		// this is a player
		asteroidTag := s.AsteroidTags.Get(e2)
		if asteroidTag != nil {
			hp := s.Hps.Get(e1)

			hp.Hp -= 1

			sfxEntity := s.EntityManager.Create()

			playerPos := s.Positions.Get(e1)
			s.Positions.Create(sfxEntity, stdcomponents.Position{XY: playerPos.XY})

			s.SoundEffects.Create(sfxEntity, components.SoundEffect{
				Clip:      assets.Audio.Get("damage_sound.wav"),
				Pitch:     float32(1.0 + (float32(hp.MaxHp)-float32(hp.Hp))/float32(hp.MaxHp)), // higher pitch = less hp
				IsPlaying: false,
				IsLooping: false,
				Volume:    1.0,
				Pan:       0.5,
			})

			return true
		}

		//wallTag := s.WallTags.Get(e2)
		//if wallTag != nil {
		//	// reverse player movement vector
		//	velocity := s.Velocities.Get(e1)
		//	rotation := s.Rotations.Get(e1)
		//	velocity.X *= -1
		//	velocity.Y *= -1
		//	rotation.Angle += 180
		//	return true
		//}
	} else if e2Tag != nil {
		// this is a player
		asteroidTag := s.AsteroidTags.Get(e1)
		if asteroidTag != nil {
			hp := s.Hps.Get(e2)
			hp.Hp -= 1
			return true
		}

		//wallTag := s.WallTags.Get(e1)
		//if wallTag != nil {
		//	// reverse player movement vector
		//	velocity := s.Velocities.Get(e2)
		//	rotation := s.Rotations.Get(e2)
		//	velocity.X *= -1
		//	velocity.Y *= -1
		//	rotation.Angle += 180
		//	return true
		//}
	}

	return false
}

func (s *CollisionHandlerSystem) checkBulletCollisionEnter(e1, e2 ecs.Entity) bool {
	e1Tag := s.BulletTags.Get(e1)
	e2Tag := s.BulletTags.Get(e2)

	if e1Tag != nil {
		// this is a bullet
		bulletHp := s.Hps.Get(e1)
		asteroidTag := s.AsteroidTags.Get(e2)
		if asteroidTag != nil {
			asteroidHp := s.Hps.Get(e2)
			asteroidHp.Hp -= 1
			bulletHp.Hp -= 1
			return true
		}
		wallTag := s.WallTags.Get(e2)
		if wallTag != nil {
			bulletHp.Hp = 0
			return true
		}
	} else if e2Tag != nil {
		// this is a bullet
		bulletHp := s.Hps.Get(e2)
		asteroidTag := s.AsteroidTags.Get(e1)
		if asteroidTag != nil {
			asteroidHp := s.Hps.Get(e1)
			asteroidHp.Hp -= 1
			bulletHp.Hp -= 1
			return true
		}
	}

	return false
}
