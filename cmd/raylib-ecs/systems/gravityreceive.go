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

const G = 6.67430e-1

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

		cx, cy := calcGravityEffect(spat.Ent.ContainingNode(), spat.Ent)

		vel.X += float32(cx)
		vel.Y += float32(cy)

		return true
	})
}

func (s *gravityReceiveController) Destroy(world *ecs.World) {}

func calcGravityEffect(n *spatial.QuadNode[gravity.QuadNodeUserData, any], target *spatial.Entity[gravity.QuadNodeUserData, any]) (cx, cy float64) {
	ents := n.Entities()
	if ents == nil {
		return cx, cy
	}

	tx, ty := target.Position()

	for _, ent := range ents {
		if ent == nil || ent == target {
			continue
		}

		x, y := ent.Position()
		fx, fy := calc2BodyGravity(tx, ty, 1.0, x, y, 1.0)
		cx += fx
		cy += fy
	}

	cur := n.Parent()
	for cur != nil {
		if cur.Parent() != nil {
			for _, sibling := range cur.Parent().Childs() {
				if sibling == nil || sibling == cur {
					continue
				}

				mass := sibling.UserData().Mass
				x := sibling.UserData().GX
				y := sibling.UserData().GY
				fx, fy := calc2BodyGravity(tx, ty, 1.0, x, y, mass)
				cx += fx
				cy += fy
			}
		}

		cur = cur.Parent()
	}

	return cx, cy
}

func calc2BodyGravity(ax, ay, am, bx, by, bm float64) (fx, fy float64) {
	dx, dy := bx-ax, by-ay
	distSq := dx*dx + dy*dy + 2.0
	if distSq != 0.0 {
		frc := G * (am * bm / distSq)
		dist := math.Sqrt(distSq)

		fx = frc * (dx / dist)
		fy = frc * (dy / dist)
	}

	return fx, fy
}
