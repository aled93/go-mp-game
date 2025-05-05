package util

import "math"

// Mat4 is Matrix with size 4x4 with row major order
type Mat4 [4]Vec4

func NewMat4(
	r1c1, r1c2, r1c3, r1c4 Scalar,
	r2c1, r2c2, r2c3, r2c4 Scalar,
	r3c1, r3c2, r3c3, r3c4 Scalar,
	r4c1, r4c2, r4c3, r4c4 Scalar) Mat4 {
	return Mat4{
		Vec4{X: r1c1, Y: r1c2, Z: r1c3, W: r1c4},
		Vec4{X: r2c1, Y: r2c2, Z: r2c3, W: r2c4},
		Vec4{X: r3c1, Y: r3c2, Z: r3c3, W: r3c4},
		Vec4{X: r4c1, Y: r4c2, Z: r4c3, W: r4c4},
	}
}

func NewMat4Identity() Mat4 {
	return Mat4{
		Vec4{1.0, 0.0, 0.0, 0.0},
		Vec4{0.0, 1.0, 0.0, 0.0},
		Vec4{0.0, 0.0, 1.0, 0.0},
		Vec4{0.0, 0.0, 0.0, 1.0},
	}
}

func NewMat4FromRows(row1, row2, row3, row4 Vec4) Mat4 {
	return Mat4{row1, row2, row3, row4}
}

func NewMat4FromColumns(col1, col2, col3, col4 Vec4) Mat4 {
	return Mat4{
		Vec4{X: col1.X, Y: col2.X, Z: col3.X, W: col4.X},
		Vec4{X: col1.Y, Y: col2.Y, Z: col3.Y, W: col4.Y},
		Vec4{X: col1.Z, Y: col2.Z, Z: col3.Z, W: col4.Z},
		Vec4{X: col1.W, Y: col2.W, Z: col3.W, W: col4.W},
	}
}

func NewMat4FromTranslate(translate Vec3) Mat4 {
	return Mat4{
		Vec4{1.0, 0.0, 0.0, 0.0},
		Vec4{0.0, 1.0, 0.0, 0.0},
		Vec4{0.0, 0.0, 1.0, 0.0},
		Vec4{translate.X, translate.Y, translate.Z, 1.0},
	}
}

func NewMat4FromScale(scale Vec3) Mat4 {
	return Mat4{
		Vec4{scale.X, 0.0, 0.0, 0.0},
		Vec4{0.0, scale.Y, 0.0, 0.0},
		Vec4{0.0, 0.0, scale.Z, 0.0},
		Vec4{0.0, 0.0, 0.0, 1.0},
	}
}

func NewMat4FromEuler(angles Vec3) Mat4 {
	cr := Scalar(math.Cos(float64(angles.X)))
	sr := Scalar(math.Sin(float64(angles.X)))
	cp := Scalar(math.Cos(float64(angles.Y)))
	sp := Scalar(math.Sin(float64(angles.Y)))
	cy := Scalar(math.Cos(float64(angles.Z)))
	sy := Scalar(math.Sin(float64(angles.Z)))

	srsp := sr * sp
	crsp := cr * sp

	return Mat4{
		Vec4{cp * cy, cp * sy, -sp, 0.0},
		Vec4{srsp*cy - cr*sy, srsp*sy + cr*cy, sr * cp, 0.0},
		Vec4{crsp*cy + sr*sy, crsp*sy - sr*cy, cr * cp, 0.0},
		Vec4{0.0, 0.0, 0.0, 1.0},
	}
}

// TODO: NewMat4FromMat3
// TODO: NewMat4FromQuaternion
// TODO: NewMat4FromOrthogonal
// TODO: NewMat4FromPerspective

// NewMat4FromAxisAngle creates rotation matrix which rotates around axis
// by given angle in radians
func NewMat4FromAxisAngle(axis Vec3, angle Scalar) Mat4 {
	x := axis.X
	y := axis.Y
	z := axis.Z

	sn := Scalar(math.Sin(float64(angle)))
	cs := Scalar(math.Cos(float64(angle)))

	xx := x * x
	yy := y * y
	zz := z * z
	xy := x * y
	xz := x * z
	yz := y * z

	return Mat4{
		Vec4{
			xx + (cs * (1.0 - xx)),
			(xy - (cs * xy)) - (sn * z),
			(xz - (cs * xz)) + (sn * y),
			0.0,
		},
		Vec4{
			(xy - (cs * xy)) + (sn * z),
			yy + (cs * (1.0 - yy)),
			(yz - (cs * yz)) - (sn * x),
			0.0,
		},
		Vec4{
			(xz - (cs * xz)) - (sn * y),
			(yz - (cs * yz)) + (sn * x),
			zz + (cs * (1.0 - zz)),
			0.0,
		},
		Vec4{0.0, 0.0, 0.0, 1.0},
	}
}

func (m *Mat4) Row0() Vec4 {
	return m[0]
}

func (m *Mat4) Row1() Vec4 {
	return m[1]
}

func (m *Mat4) Row2() Vec4 {
	return m[2]
}

func (m *Mat4) Row3() Vec4 {
	return m[3]
}

func (m *Mat4) Column0() Vec4 {
	return Vec4{m[0].X, m[1].X, m[2].X, m[3].X}
}

func (m *Mat4) Column1() Vec4 {
	return Vec4{m[0].Y, m[1].Y, m[2].Y, m[3].Y}
}

func (m *Mat4) Column2() Vec4 {
	return Vec4{m[0].Z, m[1].Z, m[2].Z, m[3].Z}
}

func (m *Mat4) Column3() Vec4 {
	return Vec4{m[0].W, m[1].W, m[2].W, m[3].W}
}

func (m *Mat4) Transpose() Mat4 {
	return Mat4{m.Column0(), m.Column1(), m.Column2(), m.Column3()}
}

func (m *Mat4) Deterninant() Scalar {
	return m[0].X*m[1].Y*m[2].Z*m[3].W - m[0].X*m[1].Y*m[2].W*m[3].Z + m[0].X*m[1].Z*m[2].W*m[3].Y - m[0].X*m[1].Z*m[2].Y*m[3].W +
		m[0].X*m[1].W*m[2].Y*m[3].Z - m[0].X*m[1].W*m[2].Z*m[3].Y - m[0].Y*m[1].Z*m[2].W*m[3].X + m[0].Y*m[1].Z*m[2].X*m[3].W -
		m[0].Y*m[1].W*m[2].X*m[3].Z + m[0].Y*m[1].W*m[2].Z*m[3].X - m[0].Y*m[1].X*m[2].Z*m[3].W + m[0].Y*m[1].X*m[2].W*m[3].Z +
		m[0].Z*m[1].W*m[2].X*m[3].Y - m[0].Z*m[1].W*m[2].Y*m[3].X + m[0].Z*m[1].X*m[2].Y*m[3].W - m[0].Z*m[1].X*m[2].W*m[3].Y +
		m[0].Z*m[1].Y*m[2].W*m[3].X - m[0].Z*m[1].Y*m[2].X*m[3].W - m[0].W*m[1].X*m[2].Y*m[3].Z + m[0].W*m[1].X*m[2].Z*m[3].Y -
		m[0].W*m[1].Y*m[2].Z*m[3].X + m[0].W*m[1].Y*m[2].X*m[3].Z - m[0].W*m[1].Z*m[2].X*m[3].Y + m[0].W*m[1].Z*m[2].Y*m[3].X
}

func (m Mat4) Multiply(o Mat4) Mat4 {
	return NewMat4(
		m[0].X*o[0].X+m[1].X*o[0].Y+m[2].X*o[0].Z+m[3].X*o[0].W,
		m[0].Y*o[0].X+m[1].Y*o[0].Y+m[2].Y*o[0].Z+m[3].Y*o[0].W,
		m[0].Z*o[0].X+m[1].Z*o[0].Y+m[2].Z*o[0].Z+m[3].Z*o[0].W,
		m[0].W*o[0].X+m[1].W*o[0].Y+m[2].W*o[0].Z+m[3].W*o[0].W,

		m[0].X*o[1].X+m[1].X*o[1].Y+m[2].X*o[1].Z+m[3].X*o[1].W,
		m[0].Y*o[1].X+m[1].Y*o[1].Y+m[2].Y*o[1].Z+m[3].Y*o[1].W,
		m[0].Z*o[1].X+m[1].Z*o[1].Y+m[2].Z*o[1].Z+m[3].Z*o[1].W,
		m[0].W*o[1].X+m[1].W*o[1].Y+m[2].W*o[1].Z+m[3].W*o[1].W,

		m[0].X*o[2].X+m[1].X*o[2].Y+m[2].X*o[2].Z+m[3].X*o[2].W,
		m[0].Y*o[2].X+m[1].Y*o[2].Y+m[2].Y*o[2].Z+m[3].Y*o[2].W,
		m[0].Z*o[2].X+m[1].Z*o[2].Y+m[2].Z*o[2].Z+m[3].Z*o[2].W,
		m[0].W*o[2].X+m[1].W*o[2].Y+m[2].W*o[2].Z+m[3].W*o[2].W,

		m[0].X*o[3].X+m[1].X*o[3].Y+m[2].X*o[3].Z+m[3].X*o[3].W,
		m[0].Y*o[3].X+m[1].Y*o[3].Y+m[2].Y*o[3].Z+m[3].Y*o[3].W,
		m[0].Z*o[3].X+m[1].Z*o[3].Y+m[2].Z*o[3].Z+m[3].Z*o[3].W,
		m[0].W*o[3].X+m[1].W*o[3].Y+m[2].W*o[3].Z+m[3].W*o[3].W,
	)
}

func (m Mat4) MultiplyScalar(o Scalar) Mat4 {
	return NewMat4(
		m[0].X*o, m[0].Y*o, m[0].Z*o, m[0].W*o,
		m[1].X*o, m[1].Y*o, m[1].Z*o, m[1].W*o,
		m[2].X*o, m[2].Y*o, m[2].Z*o, m[2].W*o,
		m[3].X*o, m[3].Y*o, m[3].Z*o, m[3].W*o,
	)
}

func (m Mat4) TransformVec4(o Vec4) Mat4 {
	return NewMat4(
		m[0].X*o.X+m[1].X*o.Y+m[2].X*o.Z+m[3].X*o.W,
		m[0].Y*o.X+m[1].Y*o.Y+m[2].Y*o.Z+m[3].Y*o.W,
		m[0].Z*o.X+m[1].Z*o.Y+m[2].Z*o.Z+m[3].Z*o.W,
		m[0].W*o.X+m[1].W*o.Y+m[2].W*o.Z+m[3].W*o.W,

		m[0].X*o.X+m[1].X*o.Y+m[2].X*o.Z+m[3].X*o.W,
		m[0].Y*o.X+m[1].Y*o.Y+m[2].Y*o.Z+m[3].Y*o.W,
		m[0].Z*o.X+m[1].Z*o.Y+m[2].Z*o.Z+m[3].Z*o.W,
		m[0].W*o.X+m[1].W*o.Y+m[2].W*o.Z+m[3].W*o.W,

		m[0].X*o.X+m[1].X*o.Y+m[2].X*o.Z+m[3].X*o.W,
		m[0].Y*o.X+m[1].Y*o.Y+m[2].Y*o.Z+m[3].Y*o.W,
		m[0].Z*o.X+m[1].Z*o.Y+m[2].Z*o.Z+m[3].Z*o.W,
		m[0].W*o.X+m[1].W*o.Y+m[2].W*o.Z+m[3].W*o.W,

		m[0].X*o.X+m[1].X*o.Y+m[2].X*o.Z+m[3].X*o.W,
		m[0].Y*o.X+m[1].Y*o.Y+m[2].Y*o.Z+m[3].Y*o.W,
		m[0].Z*o.X+m[1].Z*o.Y+m[2].Z*o.Z+m[3].Z*o.W,
		m[0].W*o.X+m[1].W*o.Y+m[2].W*o.Z+m[3].W*o.W,
	)
}
