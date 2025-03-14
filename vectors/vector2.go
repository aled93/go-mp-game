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

package vectors

import "math"

type Vec2 struct {
	X, Y float32
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

func (v Vec2) Mul(other Vec2) Vec2 {
	return Vec2{v.X * other.X, v.Y * other.Y}
}

func (v Vec2) Div(other Vec2) Vec2 {
	return Vec2{v.X / other.X, v.Y / other.Y}
}

func (v Vec2) AddScalar(scalar float32) Vec2 {
	return Vec2{v.X + scalar, v.Y + scalar}
}

func (v Vec2) SubScalar(scalar float32) Vec2 {
	return Vec2{v.X - scalar, v.Y - scalar}
}

func (v Vec2) Scale(scalar float32) Vec2 {
	return Vec2{v.X * scalar, v.Y * scalar}
}

func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v Vec2) Normalize() Vec2 {
	return v.Scale(1 / v.Length())
}
