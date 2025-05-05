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
)

type Simplex2d struct {
	a, b, c util.Vec3
	count   int
}

func (s *Simplex2d) do(direction *util.Vec2) bool {
	ao := s.a.Back()

	switch s.count {
	case 2: // Line
		ab := s.b.Subtract(s.a)
		if ab.Dot(ao) > 0 {
			newDirection := ab.Cross(ao).Cross(ab)
			direction.X = newDirection.X
			direction.Y = newDirection.Y
		} else {
			direction.X = ao.X
			direction.Y = ao.Y
			s.count = 1
		}
	case 3: // Triangle
		ab := s.b.Subtract(s.a)
		ac := s.c.Subtract(s.a)
		abc := ab.Cross(ac)

		if abc.Cross(ac).Dot(ao) > 0 {
			if ac.Dot(ao) > 0 {
				newDirection := ac.Cross(ao).Cross(ac)
				direction.X = newDirection.X
				direction.Y = newDirection.Y
				s.b = s.c
				s.count = 2
			} else {
				if ab.Dot(ao) > 0 {
					newDirection := ab.Cross(ao).Cross(ab)
					direction.X = newDirection.X
					direction.Y = newDirection.Y
					s.count = 2
				} else {
					direction.X = ao.X
					direction.Y = ao.Y
					s.count = 1
				}
			}
		} else {
			if ab.Cross(abc).Dot(ao) > 0 {
				if ab.Dot(ao) > 0 {
					newDirection := ab.Cross(ao).Cross(ab)
					direction.X = newDirection.X
					direction.Y = newDirection.Y
					s.count = 2
				} else {
					direction.X = ao.X
					direction.Y = ao.Y
					s.count = 1
				}
			} else {
				return true
				// if abc.Dot(ao) > 0 {
				// 	newDirection := abc
				// 	direction.X = newDirection.X
				// 	direction.Y = newDirection.Y
				// } else {
				// 	s.b, s.c = s.a, s.b
				// 	newDirection := abc.Neg()
				// 	direction.X = newDirection.X
				// 	direction.Y = newDirection.Y
				// }
			}
		}
	default:
		panic("Invalid simplex")
	}
	return false
}

func (s *Simplex2d) add(p util.Vec3) {
	switch s.count {
	case 0:
		s.a = p
	case 1:
		s.a, s.b = p, s.a
	case 2:
		s.a, s.b, s.c = p, s.a, s.b
	default:
		panic("Invalid simplex")
	}
	s.count++
}

func (s *Simplex2d) toPolytope(polytope []util.Vec2) []util.Vec2 {
	switch s.count {
	case 1:
		polytope = append(polytope, util.NewVec2(s.a.X, s.a.Y))
	case 2:
		polytope = append(polytope, util.NewVec2(s.a.X, s.a.Y), util.NewVec2(s.b.X, s.b.Y))
	case 3:
		polytope = append(polytope, util.NewVec2(s.a.X, s.a.Y), util.NewVec2(s.b.X, s.b.Y), util.NewVec2(s.c.X, s.c.Y))
	default:
		panic("Invalid simplex")
	}
	return polytope
}
