package util

import "math"

type Vec4 struct {
	X, Y, Z, W Scalar
}

func NewVec4(x, y, z, w Scalar) Vec4 {
	return Vec4{
		X: x,
		Y: y,
		Z: z,
		W: w,
	}
}

func NewVec4FromScalar(scalar Scalar) Vec4 {
	return Vec4{
		X: scalar,
		Y: scalar,
		Z: scalar,
		W: scalar,
	}
}

func NewVec4FromVec2(vec2 Vec2, z, w Scalar) Vec4 {
	return Vec4{
		X: vec2.X,
		Y: vec2.Y,
		Z: z,
		W: w,
	}
}

func NewVec4FromVec3(vec3 Vec3, w Scalar) Vec4 {
	return Vec4{
		X: vec3.X,
		Y: vec3.Y,
		Z: vec3.Z,
		W: w,
	}
}

func (v Vec4) Add(other Vec4) Vec4 {
	return NewVec4(v.X+other.X, v.Y+other.Y, v.Z+other.Z, v.W+other.W)
}

func (v Vec4) AddScalar(other Scalar) Vec4 {
	return NewVec4(v.X+other, v.Y+other, v.Z+other, v.W+other)
}

func (v Vec4) Subtract(other Vec4) Vec4 {
	return NewVec4(v.X-other.X, v.Y-other.Y, v.Z-other.Z, v.W-other.W)
}

func (v Vec4) SubtractScalar(other Scalar) Vec4 {
	return NewVec4(v.X-other, v.Y-other, v.Z-other, v.W-other)
}

func (v Vec4) Scale(other Vec4) Vec4 {
	return NewVec4(v.X*other.X, v.Y*other.Y, v.Z*other.Z, v.W*other.W)
}

func (v Vec4) ScaleScalar(other Scalar) Vec4 {
	return NewVec4(v.X*other, v.Y*other, v.Z*other, v.W*other)
}

func (v Vec4) Divide(other Vec4) Vec4 {
	return NewVec4(v.X/other.X, v.Y/other.Y, v.Z/other.Z, v.W/other.W)
}

func (v Vec4) DivideScalar(other Scalar) Vec4 {
	return NewVec4(v.X/other, v.Y/other, v.Z/other, v.W/other)
}

func (v Vec4) LengthSquare() Scalar {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

func (v Vec4) Length() Scalar {
	return Scalar(math.Sqrt(float64(v.LengthSquare())))
}

func (v Vec4) Normalized() Vec4 {
	return v.DivideScalar(v.Length())
}

func (v Vec4) DistanceSquare(to Vec4) Scalar {
	return v.Subtract(to).LengthSquare()
}

func (v Vec4) Distance(to Vec4) Scalar {
	return v.Subtract(to).Length()
}

// Lerp interpolate vector between v and target. With fraction == 0.0
// result vector will be equals to v, with fraction == 1.0 result will
// be equals target.
func (v Vec4) Lerp(target Vec4, fraction Scalar) Vec4 {
	return v.Add(target.Subtract(v).ScaleScalar(fraction))
}

// Min returns vector with each component having lowest values of v and other
func (v Vec4) Min(other Vec4) Vec4 {
	return NewVec4(min(v.X, other.X), min(v.Y, other.Y), min(v.Z, other.Z), min(v.W, other.W))
}

// Max returns vector with each component having highest values of v and other
func (v Vec4) Max(other Vec4) Vec4 {
	return NewVec4(max(v.X, other.X), max(v.Y, other.Y), max(v.Z, other.Z), max(v.W, other.W))
}

func (v Vec4) Dot(other Vec4) Scalar {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z + v.W*other.W
}

func (v Vec4) Back() Vec4 {
	return NewVec4(-v.X, -v.Y, -v.Z, -v.W)
}
