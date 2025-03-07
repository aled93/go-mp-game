package spatial

import (
	"iter"
)

type AABB2D struct {
	XMin, YMin float64
	XMax, YMax float64
}

func NewAABBFromCenterSize(ox, oy, w, h float64) AABB2D {
	hw := w * 0.5
	hh := h * 0.5
	return AABB2D{
		XMin: ox - hw,
		YMin: oy - hh,
		XMax: ox + hw,
		YMax: oy + hh,
	}
}

func (aabb *AABB2D) Contains(x, y float64) bool {
	return x >= aabb.XMin && x < aabb.XMax &&
		y >= aabb.YMin && y < aabb.YMax
}

func (aabb *AABB2D) Intersects(aabb2 *AABB2D) bool {
	return !(aabb.XMax < aabb2.XMin ||
		aabb.XMin > aabb2.XMax ||
		aabb.YMax < aabb2.YMin ||
		aabb.YMin > aabb2.YMax)
}

func (aabb *AABB2D) Shift(dx, dy float64) {
	aabb.XMin += dx
	aabb.YMin += dy
	aabb.XMax += dx
	aabb.YMax += dy
}

type Partitioner2D[T any] interface {
	Register(data T, bounds AABB2D) SpatialObject[T]
	Unregister(object SpatialObject[T])
	UpdateObject(object SpatialObject[T], newData T, newBounds AABB2D)

	QueryPointFunc(x, y float64, f func(SpatialObject[T]) bool) (SpatialObject[T], bool)
	QueryPointFirst(x, y float64) (SpatialObject[T], bool)
	QueryPointIter(x, y float64) iter.Seq[SpatialObject[T]]
	QueryAreaIter(area AABB2D) iter.Seq[SpatialObject[T]]
	QueryAreaAll(area AABB2D, out []SpatialObject[T]) []SpatialObject[T]
}

type FinitePartitioner2D[T any] interface {
	Partitioner2D[T]
	Bounds() AABB2D
}

type SpatialObject[T any] interface {
	GetData() T
}
