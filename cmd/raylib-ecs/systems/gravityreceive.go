/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
	"gomp_game/pkgs/spatial"
	"math"
)

const G = 6.67430e-11

type gravityReceiveController struct {
	spatials   *ecs.ComponentManager[components.SpatialEnt]
	gravRecvs  *ecs.ComponentManager[components.GravitationReceiver]
	velocities *ecs.ComponentManager[components.Velocity]
}

func (s *gravityReceiveController) Init(world *ecs.World) {
	s.spatials = components.SpatialEntService.GetManager(world)
	s.gravRecvs = components.GravitationReceiveService.GetManager(world)
	s.velocities = components.VelocityService.GetManager(world)
}

func (s *gravityReceiveController) Update(world *ecs.World) {}

func (s *gravityReceiveController) FixedUpdate(world *ecs.World) {
	s.spatials.AllParallel(func(entId ecs.EntityID, spat *components.SpatialEnt) bool {
		if !s.gravRecvs.Has(entId) {
			return true
		}

		vel := s.velocities.Get(entId)
		if vel == nil {
			return true
		}

		cx, cy, mass := calcGravityEffect(spat.Ent.ContainingNode(), spat.Ent)
		x, y := spat.Ent.Position()
		dx := cx - x
		dy := cy - y
		distSq := dx*dx + dy*dy

		if distSq == 0.0 {
			return true
		}

		frc := mass / distSq
		dist := math.Sqrt(distSq)

		vel.X += float32(frc * (dx / dist))
		vel.Y += float32(frc * (dy / dist))

		return true
	})
}

func (s *gravityReceiveController) Destroy(world *ecs.World) {}

func calcGravityEffect(n *spatial.QuadNode[gravity.QuadNodeUserData, any], target *spatial.Entity[gravity.QuadNodeUserData, any]) (cx, cy, mass float64) {
	ents := n.Entities()
	if ents == nil {
		return cx, cy, mass
	}

	ei := 0
	for _, ent := range ents {
		if ent == nil || ent == target {
			continue
		}

		if ei == 0 {
			cx, cy = ent.Position()
		} else {
			x, y := ent.Position()
			cx += x
			cy += y
			mass += 1.0
		}
		ei++
	}

	cur := n.Parent()
	for cur != nil {
		if cur.Parent() != nil {
			for _, sibling := range cur.Parent().Childs() {
				if sibling == nil || sibling == cur {
					continue
				}

				mass += sibling.UserData().Mass
				cx += sibling.UserData().GX * sibling.UserData().Mass
				cy += sibling.UserData().GY * sibling.UserData().Mass
			}
		}

		cur = cur.Parent()
	}

	return cx / mass, cy / mass, mass
}
