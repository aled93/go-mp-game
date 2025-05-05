package util

import "math"

type Vec2 struct {
	X, Y Scalar
}

func NewVec2(x, y Scalar) Vec2 {
	return Vec2{
		X: x,
		Y: y,
	}
}

func NewVec2FromScalar(scalar Scalar) Vec2 {
	return Vec2{
		X: scalar,
		Y: scalar,
	}
}

func (v Vec2) Add(other Vec2) Vec2 {
	return NewVec2(v.X+other.X, v.Y+other.Y)
}

func (v Vec2) AddScalar(other Scalar) Vec2 {
	return NewVec2(v.X+other, v.Y+other)
}

func (v Vec2) Subtract(other Vec2) Vec2 {
	return NewVec2(v.X-other.X, v.Y-other.Y)
}

func (v Vec2) SubtractScalar(other Scalar) Vec2 {
	return NewVec2(v.X-other, v.Y-other)
}

func (v Vec2) Scale(other Vec2) Vec2 {
	return NewVec2(v.X*other.X, v.Y*other.Y)
}

func (v Vec2) ScaleScalar(other Scalar) Vec2 {
	return NewVec2(v.X*other, v.Y*other)
}

func (v Vec2) Divide(other Vec2) Vec2 {
	return NewVec2(v.X/other.X, v.Y/other.Y)
}

func (v Vec2) DivideScalar(other Scalar) Vec2 {
	return NewVec2(v.X/other, v.Y/other)
}

func (v Vec2) LengthSquared() Scalar {
	return v.X*v.X + v.Y*v.Y
}

func (v Vec2) Length() Scalar {
	return Scalar(math.Sqrt(float64(v.LengthSquared())))
}

func (v Vec2) Normalized() Vec2 {
	return v.DivideScalar(v.Length())
}

func (v Vec2) DistanceSquared(to Vec2) Scalar {
	return v.Subtract(to).LengthSquared()
}

func (v Vec2) Distance(to Vec2) Scalar {
	return v.Subtract(to).Length()
}

// Angle returns angle of v vector in radians.
func (v Vec2) Angle() Scalar {
	return Scalar(math.Atan2(float64(v.Y), float64(v.X)))
}

// AngleTo returns angle of vector with origin at v and
// ending at to.
func (v Vec2) AngleTo(to Vec2) Scalar {
	return to.Subtract(v).Angle()
}

func (v Vec2) Rotate(angle Radians) Vec2 {
	return NewVec2(
		v.X*float32(math.Cos(angle))-v.Y*float32(math.Sin(angle)),
		v.X*float32(math.Sin(angle))+v.Y*float32(math.Cos(angle)),
	)
}

// Lerp interpolate vector between v and target. With fraction == 0.0
// result vector will be equals to v, with fraction == 1.0 result will
// be equals target.
func (v Vec2) Lerp(target Vec2, fraction Scalar) Vec2 {
	return v.Add(target.Subtract(v).ScaleScalar(fraction))
}

// Min returns vector with each component having lowest values of v and other
func (v Vec2) Min(other Vec2) Vec2 {
	return NewVec2(min(v.X, other.X), min(v.Y, other.Y))
}

// Max returns vector with each component having highest values of v and other
func (v Vec2) Max(other Vec2) Vec2 {
	return NewVec2(max(v.X, other.X), max(v.Y, other.Y))
}

// Back returns opposite direction to v
func (v Vec2) Back() Vec2 {
	return NewVec2(-v.X, -v.Y)
}

// PerpendicularCW returns direction rotated 90 degrees clockwise
func (v Vec2) PerpendicularCW() Vec2 {
	return NewVec2(v.Y, -v.X)
}

func (v Vec2) Dot(other Vec2) Scalar {
	return v.X*other.X + v.Y*other.Y
}
