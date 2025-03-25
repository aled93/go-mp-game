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

func minkowskiSupport2d(a, b AnyCollider, transformA, transformB *stdcomponents.Transform2d, direction vectors.Vec2) vectors.Vec2 {
	return a.GetSupport(direction, transformA).Sub(b.GetSupport(direction.Neg(), transformB))
}
