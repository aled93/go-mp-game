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
	"math/rand"
	"time"
)

func NewHpSystem() HpSystem {
	return HpSystem{}
}

type HpSystem struct {
	EntityManager        *ecs.EntityManager
	Hps                  *components.HpComponentManager
	Asteroids            *components.AsteroidComponentManager
	Players              *components.PlayerTagComponentManager
	Hp                   *components.HpComponentManager
	AsteroidSceneManager *components.AsteroidSceneManagerComponentManager
	Positions            *stdcomponents.PositionComponentManager
	Velocities           *stdcomponents.VelocityComponentManager
	BoxColliders         *stdcomponents.BoxColliderComponentManager
	Sprites              *stdcomponents.SpriteComponentManager
	Pickups              *components.PickupComponentManager
}

func (s *HpSystem) Init() {}
func (s *HpSystem) Run(dt time.Duration) {
	s.Hps.EachEntity(func(e ecs.Entity) bool {
		hp := s.Hps.Get(e)

		if hp.Hp <= 0 {
			asteroid := s.Asteroids.Get(e)
			if asteroid != nil && rand.Intn(10) == 0 {
				pos := s.Positions.Get(e)
				vel := s.Velocities.Get(e)

				if pos != nil && vel != nil {
					entities.CreatePickup(entities.PickupManagers{
						EntityManager: s.EntityManager,
						Positions:     s.Positions,
						Velocities:    s.Velocities,
						BoxColliders:  s.BoxColliders,
						Sprites:       s.Sprites,
						Pickups:       s.Pickups,
					}, pos.X, pos.Y, vel.X, vel.Y, components.Pickup{
						Power:  components.PickupPower_Hp,
						Amount: 1,
					})
				}
			}

			player := s.Players.Get(e)
			s.AsteroidSceneManager.EachComponent(func(a *components.AsteroidSceneManager) bool {
				if asteroid != nil {
					a.PlayerScore += hp.MaxHp
				}
				if player != nil {
					playerHp := s.Hp.Get(e)
					a.PlayerHp = playerHp.Hp
				}
				return false
			})

			s.EntityManager.Delete(e)
		}
		return true
	})
}
func (s *HpSystem) Destroy() {}
