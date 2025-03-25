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

package gjk

import (
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
)

func EPA(
	a, b AnyCollider,
	transformA, transformB *stdcomponents.Transform2d,
	simplex *Simplex2d,
) (vectors.Vec2, float32) {
	polytope := simplex.toPolytope(make([]vectors.Vec2, 0, 6))

	for range maxItterations {
		edge := findClosestEdge(polytope)
		point := minkowskiSupport2d(a, b, transformA, transformB, edge.normal)
		distance := point.Dot(edge.normal)
		if distance-edge.distance < epaTolerance {
			return edge.normal, distance
		}

		polytope = append(polytope[:edge.index], append([]vectors.Vec2{point}, polytope[edge.index:]...)...)
	}

	panic("EPA infinite loop")
}

func findClosestEdge(polytope []vectors.Vec2) closestEdge {
	closest := closestEdge{
		distance: float32(math.MaxFloat32),
		normal:   vectors.Vec2{},
		index:    -1,
	}

	for i := 0; i < len(polytope); i++ {
		j := (i + 1) % len(polytope)
		a := polytope[i]
		b := polytope[j]

		edge := b.Sub(a).ToVec3()
		oa := a.ToVec3()

		normal := edge.Cross(oa).Cross(edge).ToVec2().Normalize()
		distance := normal.Dot(a)

		if distance < closest.distance {
			closest.distance = distance
			closest.normal = normal
			closest.index = j
		}
	}

	return closest
}

type closestEdge struct {
	distance float32
	normal   vectors.Vec2
	index    int
}
