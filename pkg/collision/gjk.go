/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

Thank you for your support!
*/

package gjk

import (
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
)

const (
	maxItterations = 64
	epaTolerance   = 0.001
)

type AnyCollider interface {
	GetSupport(direction vectors.Vec2, transform *stdcomponents.Transform2d) vectors.Vec2
}

func CheckCollision(
	a, b AnyCollider,
	transformA, transformB *stdcomponents.Transform2d,
) (Simplex2d, bool) {
	direction := vectors.Vec2{X: 1, Y: 0}
	simplex := Simplex2d{}

	p := minkowskiSupport2d(a, b, transformA, transformB, direction)
	simplex.add(p.ToVec3())
	direction = p.Neg()

	for range maxItterations {
		p = minkowskiSupport2d(a, b, transformA, transformB, direction)

		if p.Dot(direction) < 0 {
			return simplex, false
		}

		simplex.add(p.ToVec3())

		if simplex.do(&direction) {
			return simplex, true
		}
	}

	panic("GJK infinite loop")
}

func EPA(
	a, b AnyCollider,
	transformA, transformB *stdcomponents.Transform2d,
	simplex *Simplex2d,
) (vectors.Vec2, float32) {
	var minIndex int = 0
	var minDistance float32 = float32(math.MaxFloat32)
	var minNormal vectors.Vec2
	var polytope = simplex.toPolytope(make([]vectors.Vec2, 0, 6))

	for range maxItterations {
		for i := 0; i < len(polytope); i++ {
			j := (i + 1) % len(polytope)
			a := polytope[i]
			b := polytope[j]

			edge := b.Sub(a)

			// normal := edge.Cross(a.ToVec3()).Cross(edge).ToVec2().Normalize()
			normal := edge.Normal().Normalize()
			distance := a.Dot(normal)

			if distance < 0 {
				normal = normal.Neg()
				distance = -distance
			}

			if distance < minDistance {
				minDistance = distance
				minNormal = normal
				minIndex = j
			}
		}

		if minDistance == 0 {
			return minNormal, minDistance
		}

		minNormal = minNormal.Normalize()

		p := minkowskiSupport2d(a, b, transformA, transformB, minNormal)
		normalDot := minNormal.Dot(p)
		accuracy := math.Abs(float64(normalDot - minDistance))

		if accuracy < epaTolerance {
			return minNormal, minDistance
		}

		polytope = append(polytope[:minIndex], append([]vectors.Vec2{p}, polytope[minIndex:]...)...)
	}

	panic("EPA infinite loop")
}

func minkowskiSupport2d(a, b AnyCollider, transformA, transformB *stdcomponents.Transform2d, direction vectors.Vec2) vectors.Vec2 {
	return a.GetSupport(direction, transformA).Sub(b.GetSupport(direction.Neg(), transformB))
}
