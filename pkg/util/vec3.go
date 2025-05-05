package util

import "math"

type Vec3 struct {
	X, Y, Z Scalar
}

func NewVec3(x, y, z Scalar) Vec3 {
	return Vec3{
		X: x,
		Y: y,
		Z: z,
	}
}

func NewVec3FromScalar(scalar Scalar) Vec3 {
	return Vec3{
		X: scalar,
		Y: scalar,
		Z: scalar,
	}
}

func NewVec3FromVec2(vec2 Vec2, z Scalar) Vec3 {
	return Vec3{
		X: vec2.X,
		Y: vec2.Y,
		Z: z,
	}
}

func (v Vec3) Add(other Vec3) Vec3 {
	return NewVec3(v.X+other.X, v.Y+other.Y, v.Z+other.Z)
}

func (v Vec3) AddScalar(other Scalar) Vec3 {
	return NewVec3(v.X+other, v.Y+other, v.Z+other)
}

func (v Vec3) Subtract(other Vec3) Vec3 {
	return NewVec3(v.X-other.X, v.Y-other.Y, v.Z-other.Z)
}

func (v Vec3) SubtractScalar(other Scalar) Vec3 {
	return NewVec3(v.X-other, v.Y-other, v.Z-other)
}

func (v Vec3) Scale(other Vec3) Vec3 {
	return NewVec3(v.X*other.X, v.Y*other.Y, v.Z*other.Z)
}

func (v Vec3) ScaleScalar(other Scalar) Vec3 {
	return NewVec3(v.X*other, v.Y*other, v.Z*other)
}

func (v Vec3) Divide(other Vec3) Vec3 {
	return NewVec3(v.X/other.X, v.Y/other.Y, v.Z/other.Z)
}

func (v Vec3) DivideScalar(other Scalar) Vec3 {
	return NewVec3(v.X/other, v.Y/other, v.Z/other)
}

func (v Vec3) LengthSquare() Scalar {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Length() Scalar {
	return Scalar(math.Sqrt(float64(v.LengthSquare())))
}

func (v Vec3) Normalized() Vec3 {
	return v.DivideScalar(v.Length())
}

func (v Vec3) DistanceSquare(to Vec3) Scalar {
	return v.Subtract(to).LengthSquare()
}

func (v Vec3) Distance(to Vec3) Scalar {
	return v.Subtract(to).Length()
}

// Lerp interpolate vector between v and target. With fraction == 0.0
// result vector will be equals to v, with fraction == 1.0 result will
// be equals target.
func (v Vec3) Lerp(target Vec3, fraction Scalar) Vec3 {
	return v.Add(target.Subtract(v).ScaleScalar(fraction))
}

// Min returns vector with each component having lowest values of v and other
func (v Vec3) Min(other Vec3) Vec3 {
	return NewVec3(min(v.X, other.X), min(v.Y, other.Y), min(v.Z, other.Z))
}

// Max returns vector with each component having highest values of v and other
func (v Vec3) Max(other Vec3) Vec3 {
	return NewVec3(max(v.X, other.X), max(v.Y, other.Y), max(v.Z, other.Z))
}

func (v Vec3) Dot(other Vec3) Scalar {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vec3) Cross(other Vec3) Vec3 {
	return NewVec3(
		v.Y*other.Z-v.Z*other.Y,
		v.Z*other.X-v.X*other.Z,
		v.X*other.Y-v.Y*other.X,
	)
}

func (v Vec3) Back() Vec3 {
	return NewVec3(-v.X, -v.Y, -v.Z)
}
