/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
	"gomp_game/pkgs/spatial"
)

type gravityEmitController struct{}

func (s *gravityEmitController) Init(world *ecs.World)   {}
func (s *gravityEmitController) Update(world *ecs.World) {}

func (s *gravityEmitController) FixedUpdate(world *ecs.World) {
	calcGravity(gravity.QTree.Root())
}

func (s *gravityEmitController) Destroy(world *ecs.World) {}

func calcGravity(n *spatial.QuadNode[gravity.QuadNodeUserData, any]) (cx, cy, mass float64) {
	if ents := n.Entities(); ents != nil {
		for _, ent := range ents {
			if ent == nil {
				continue
			}

			x, y := ent.Position()
			cx += x
			cy += y
			mass += 1.0
		}

		if mass != 0.0 {
			n.UserData().GX = cx / mass
			n.UserData().GY = cy / mass
		}
		n.UserData().Mass = mass

		return cx, cy, mass
	}

	for _, child := range n.Childs() {
		if child != nil {
			x, y, m := calcGravity(child)
			cx += x
			cy += y
			mass += m
		}
	}

	if mass != 0.0 {
		n.UserData().GX = cx / mass
		n.UserData().GY = cy / mass
	}
	n.UserData().Mass = mass

	return cx, cy, mass
}
