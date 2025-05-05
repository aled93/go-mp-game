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
	"gomp/pkg/util"
	"gomp/stdcomponents"
)

const (
	maxIterations = 64
)

func New() GJK {
	return GJK{}
}

type GJK struct {
	simplex Simplex2d
}

type AnyCollider interface {
	GetSupport(direction util.Vec2, transform stdcomponents.Transform2d) util.Vec2
}

/*
CheckCollision - GJK, Distance, Closest Points
https://www.youtube.com/watch?v=Qupqu1xe7Io
https://dyn4j.org/2010/04/gjk-distance-closest-points/#gjk-distance
*/
func (s *GJK) CheckCollision(
	a, b AnyCollider,
	transformA, transformB stdcomponents.Transform2d,
) bool {
	direction := util.NewVec2(1, 0)

	p := s.minkowskiSupport2d(a, b, transformA, transformB, direction)
	s.simplex.add(util.NewVec3FromVec2(p, 0))
	direction = p.Back()

	for range maxIterations {
		p = s.minkowskiSupport2d(a, b, transformA, transformB, direction)

		if p.Dot(direction) < 0 {
			return false
		}

		s.simplex.add(util.NewVec3FromVec2(p, 0))

		if s.simplex.do(&direction) {
			return true
		}
	}

	panic("GJK infinite loop")
}

func (s *GJK) minkowskiSupport2d(a, b AnyCollider, transformA, transformB stdcomponents.Transform2d, direction util.Vec2) util.Vec2 {
	aSupport := a.GetSupport(direction, transformA)
	bSupport := b.GetSupport(direction.Back(), transformB)
	support := aSupport.Subtract(bSupport)
	return support
}
