package util

type Rect struct {
	Mins, Maxs Vec2
}

func NewRect(minX, minY, maxX, maxY float32) Rect {
	return Rect{
		Mins: Vec2{
			X: minX,
			Y: minY,
		},
		Maxs: Vec2{
			X: maxX,
			Y: maxY,
		},
	}
}

func NewRectFromMinMax(mins, maxs Vec2) Rect {
	return Rect{
		Mins: mins,
		Maxs: maxs,
	}
}

func NewRectFromOriginSize(origin, size Vec2) Rect {
	return Rect{
		Mins: origin,
		Maxs: origin.Add(size),
	}
}

func NewRectFromCenterSize(center, size Vec2) Rect {
	size = size.ScaleScalar(0.5)

	return Rect{
		Mins: center.Subtract(size),
		Maxs: center.Add(size),
	}
}

func (r Rect) Center() Vec2 {
	return r.Mins.Lerp(r.Maxs, 0.5)
}

func (r Rect) Size() Vec2 {
	return r.Maxs.Subtract(r.Mins)
}

// Normalized returns normalized rectangle where r.Mins is
// lower that r.Maxs.
func (r Rect) Normalized() Rect {
	return NewRectFromMinMax(r.Mins.Min(r.Maxs), r.Mins.Max(r.Maxs))
}

// Inflate extends rectangle sides, result rectangle will be extended
// by extents*2.
func (r Rect) Inflate(extents Vec2) Rect {
	return NewRect(
		r.Mins.X-extents.X,
		r.Mins.Y-extents.Y,
		r.Maxs.X+extents.X,
		r.Maxs.Y+extents.Y,
	)
}

// Union returns rectangle containing both r and other.
func (r Rect) Union(other Rect) Rect {
	return NewRectFromMinMax(
		r.Mins.Min(r.Maxs).Min(other.Mins).Min(other.Maxs),
		r.Mins.Max(r.Maxs).Max(other.Mins).Max(other.Maxs),
	)
}

func (r Rect) ContainsPoint(point Vec2) bool {
	return point.X >= r.Mins.X &&
		point.X < r.Maxs.X &&
		point.Y >= r.Mins.Y &&
		point.Y < r.Maxs.Y
}

func (r Rect) Intersects(other Rect) bool {
	return r.Maxs.X < other.Mins.X ||
		r.Mins.X > other.Maxs.X ||
		r.Maxs.Y < other.Mins.Y ||
		r.Mins.Y > other.Maxs.Y
}
