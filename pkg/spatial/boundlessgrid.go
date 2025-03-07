package spatial

import (
	"gomp/pkg/debugdraw"
	"iter"
	"math"
	"slices"
	"time"

	"github.com/negrel/assert"
)

type BoundlessGrid2D[T any] struct {
	cellSize float64
	cells    map[[2]int64]boundlessGrid2DCell[T]
}

func NewBoundlessGrid2D[T any](cellSize float64) *BoundlessGrid2D[T] {
	assert.Greater(cellSize, 0.0, "cellSize must be positive")

	return &BoundlessGrid2D[T]{
		cellSize: cellSize,
		cells:    map[[2]int64]boundlessGrid2DCell[T]{},
	}
}

func (grid *BoundlessGrid2D[T]) register(obj *boundlessGrid2DObject[T]) {
	for y := obj.yMin; y <= obj.yMax; y++ {
		for x := obj.xMin; x <= obj.xMax; x++ {
			coords := [2]int64{x, y}
			cell, exists := grid.cells[coords]
			if !exists {
				cell = boundlessGrid2DCell[T]{}
				grid.cells[coords] = cell
			}

			cell.objects = append(cell.objects, obj)

			grid.cells[coords] = cell
		}
	}
}

// implementation of Partitioner2D

func (grid *BoundlessGrid2D[T]) Register(data T, bounds AABB2D) SpatialObject[T] {
	xMin := int64(math.Floor(bounds.XMin / grid.cellSize))
	yMin := int64(math.Floor(bounds.YMin / grid.cellSize))
	xMax := int64(math.Floor(bounds.XMax / grid.cellSize))
	yMax := int64(math.Floor(bounds.YMax / grid.cellSize))

	obj := &boundlessGrid2DObject[T]{
		data:   data,
		bounds: bounds,
		xMin:   xMin,
		yMin:   yMin,
		xMax:   xMax,
		yMax:   yMax,
	}

	grid.register(obj)

	return obj
}

func (grid *BoundlessGrid2D[T]) Unregister(object SpatialObject[T]) {
	obj, ok := object.(*boundlessGrid2DObject[T])
	if !ok {
		return
	}

	for y := obj.yMin; y <= obj.yMax; y++ {
		for x := obj.xMin; x <= obj.xMax; x++ {
			coords := [2]int64{x, y}
			if cell, exists := grid.cells[coords]; exists {
				if idx := slices.Index(cell.objects, obj); idx >= 0 {
					cell.objects = slices.Delete(cell.objects, idx, idx+1)
				}

				if len(cell.objects) > 0 {
					grid.cells[coords] = cell
				} else {
					// TODO: remove cells after some time
					delete(grid.cells, coords)
				}
			}
		}
	}
}

func (grid *BoundlessGrid2D[T]) UpdateObject(object SpatialObject[T], newData T, newBounds AABB2D) {
	obj, ok := object.(*boundlessGrid2DObject[T])
	if !ok {
		return
	}

	obj.data = newData

	xMin := int64(math.Floor(newBounds.XMin / grid.cellSize))
	yMin := int64(math.Floor(newBounds.YMin / grid.cellSize))
	xMax := int64(math.Floor(newBounds.XMax / grid.cellSize))
	yMax := int64(math.Floor(newBounds.YMax / grid.cellSize))

	if xMin == obj.xMin && yMin == obj.yMin && xMax == obj.xMax && yMax == obj.yMax {
		return
	}

	grid.Unregister(object)

	obj.bounds = newBounds
	obj.xMin = xMin
	obj.yMin = yMin
	obj.xMax = xMax
	obj.yMax = yMax

	grid.register(obj)
}

func (grid *BoundlessGrid2D[T]) QueryPointFunc(x, y float64, f func(SpatialObject[T]) bool) (SpatialObject[T], bool) {
	cellCoord := [2]int64{int64(math.Floor(x)), int64(math.Floor(y))}
	cell, exist := grid.cells[cellCoord]
	if !exist || len(cell.objects) == 0 {
		return nil, false
	}

	// without filter func
	if f == nil {
		for _, obj := range cell.objects {
			if obj.bounds.Contains(x, y) {
				return obj, true
			}
		}
		return nil, false
	}

	// with filter func
	for _, obj := range cell.objects {
		if obj.bounds.Contains(x, y) && f(obj) {
			return obj, true
		}
	}

	return nil, false
}

func (grid *BoundlessGrid2D[T]) QueryPointFirst(x, y float64) (SpatialObject[T], bool) {
	return grid.QueryPointFunc(x, y, nil)
}

func (grid *BoundlessGrid2D[T]) QueryPointIter(x, y float64) iter.Seq[SpatialObject[T]] {
	cellCoord := [2]int64{int64(math.Floor(x)), int64(math.Floor(y))}
	cell, exist := grid.cells[cellCoord]
	if !exist || len(cell.objects) == 0 {
		return func(yield func(SpatialObject[T]) bool) {}
	}

	return func(yield func(SpatialObject[T]) bool) {
		for _, obj := range cell.objects {
			if obj.bounds.Contains(x, y) && !yield(obj) {
				break
			}
		}
	}
}

func (grid *BoundlessGrid2D[T]) QueryAreaIter(area AABB2D) iter.Seq[SpatialObject[T]] {
	return func(yield func(SpatialObject[T]) bool) {
		xMin := int64(math.Floor(area.XMin / grid.cellSize))
		yMin := int64(math.Floor(area.YMin / grid.cellSize))
		xMax := int64(math.Floor(area.XMax / grid.cellSize))
		yMax := int64(math.Floor(area.YMax / grid.cellSize))

		for y := yMin; y <= yMax; y++ {
			for x := xMin; x <= xMax; x++ {
				cellCoord := [2]int64{x, y}
				cell, exist := grid.cells[cellCoord]
				if !exist {
					continue
				}

				for _, obj := range cell.objects {
					if obj.bounds.Intersects(&area) && !yield(obj) {
						return
					}
				}
			}
		}
	}
}

func (grid *BoundlessGrid2D[T]) QueryAreaAll(area AABB2D, out []SpatialObject[T]) []SpatialObject[T] {
	xMin := int64(math.Floor(area.XMin / grid.cellSize))
	yMin := int64(math.Floor(area.YMin / grid.cellSize))
	xMax := int64(math.Floor(area.XMax / grid.cellSize))
	yMax := int64(math.Floor(area.YMax / grid.cellSize))

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			cellCoord := [2]int64{x, y}
			cell, exist := grid.cells[cellCoord]
			if !exist {
				continue
			}

			for _, obj := range cell.objects {
				if obj.bounds.Intersects(&area) {
					out = append(out, obj)
				}
			}
		}
	}

	return out
}

func (grid *BoundlessGrid2D[T]) DebugDraw(duration time.Duration) {
	for coord, cell := range grid.cells {
		cx := float32(coord[0]) * float32(grid.cellSize)
		cy := float32(coord[1]) * float32(grid.cellSize)

		debugdraw.RectOutline(cx, cy, cx+float32(grid.cellSize), cy+float32(grid.cellSize), 1.0, 0.0, 0.0, 1.0, duration)

		if len(cell.objects) == 0 {
			debugdraw.RectOutline(cx+3, cy+3, cx+13, cy+13, 1.0, 0.0, 0.0, 1.0, duration)
		} else {
			for i := range len(cell.objects) {
				debugdraw.Line(cx+3+float32(i)*2, cy+3, 0.0, cx+3+float32(i)*2, cy+13, 0.0, 1.0, 0.0, 0.0, 1.0, duration)
			}
		}
	}
}

type boundlessGrid2DCell[T any] struct {
	objects []*boundlessGrid2DObject[T]
}

type boundlessGrid2DObject[T any] struct {
	data T

	bounds AABB2D

	xMin, yMin int64
	xMax, yMax int64
}

func (obj *boundlessGrid2DObject[T]) GetData() T {
	return obj.data
}
