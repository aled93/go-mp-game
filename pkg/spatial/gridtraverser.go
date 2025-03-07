package spatial

import (
	"iter"
	"math"
)

func TraverseGrid3D(from, to [3]float64) iter.Seq[GridTraverserStepResult3D] {
	return func(yield func(GridTraverserStepResult3D) bool) {
		var step GridTraverserStepResult3D
		var ok bool
		for tr := NewGridTraverser3D(from, to); ok; step, ok = tr.Step() {
			if !yield(step) {
				break
			}
		}
	}
}

type GridTraverserStepResult3D struct {
	CellX, CellY, CellZ int64
	EnterT, ExitT       float64
}

type GridTraverse3D struct {
	cur_x    int64
	cur_y    int64
	cur_z    int64
	dt_dx    float64
	dt_dy    float64
	dt_dz    float64
	t        float64
	n        int64
	x_inc    int64
	y_inc    int64
	z_inc    int64
	t_next_x float64
	t_next_y float64
	t_next_z float64
}

func NewGridTraverser3D(from, to [3]float64) GridTraverse3D {
	dx := to[0] - from[0]
	dy := to[1] - from[1]
	dz := to[2] - from[2]

	dt_dx, dt_dy, dt_dz := 0.0, 0.0, 0.0
	if dx != 0.0 {
		dt_dx = 1.0 / dx
	}
	if dy != 0.0 {
		dt_dy = 1.0 / dy
	}
	if dz != 0.0 {
		dt_dz = 1.0 / dz
	}

	tr := GridTraverse3D{
		cur_x:    int64(math.Floor(from[0])),
		cur_y:    int64(math.Floor(from[1])),
		cur_z:    int64(math.Floor(from[2])),
		dt_dx:    dt_dx,
		dt_dy:    dt_dy,
		dt_dz:    dt_dz,
		t:        0.0,
		n:        1,
		x_inc:    0,
		y_inc:    0,
		z_inc:    0,
		t_next_x: 0.0,
		t_next_y: 0.0,
		t_next_z: 0.0,
	}

	if dx == 0.0 {
		tr.x_inc = 0
		tr.t_next_x = tr.dt_dx
	} else if dx > 0.0 {
		tr.x_inc = 1
		tr.n += int64(math.Floor(to[0])) - tr.cur_x
		tr.t_next_x = (float64(tr.cur_x) + 1.0 - from[0]) * tr.dt_dx
	} else {
		tr.x_inc = -1
		tr.n += tr.cur_x - int64(math.Floor(to[0]))
		tr.t_next_x = (from[0] - float64(tr.cur_x)) * tr.dt_dx
	}

	if dy == 0.0 {
		tr.y_inc = 0
		tr.t_next_y = tr.dt_dy
	} else if dy > 0.0 {
		tr.y_inc = 1
		tr.n += int64(math.Floor(to[1])) - tr.cur_y
		tr.t_next_y = (float64(tr.cur_y) + 1.0 - from[1]) * tr.dt_dy
	} else {
		tr.y_inc = -1
		tr.n += tr.cur_y - int64(math.Floor(to[1]))
		tr.t_next_y = (from[1] - float64(tr.cur_y)) * tr.dt_dy
	}

	if dz == 0.0 {
		tr.z_inc = 0
		tr.t_next_z = tr.dt_dz
	} else if dz > 0.0 {
		tr.z_inc = 1
		tr.n += int64(math.Floor(to[2])) - tr.cur_z
		tr.t_next_z = (float64(tr.cur_z) + 1.0 - from[2]) * tr.dt_dz
	} else {
		tr.z_inc = -1
		tr.n += tr.cur_z - int64(math.Floor(to[2]))
		tr.t_next_z = (from[2] - float64(tr.cur_z)) * tr.dt_dz
	}

	return tr
}

func (tr *GridTraverse3D) Step() (stepResult GridTraverserStepResult3D, ok bool) {
	if tr.n <= 0 {
		return stepResult, false
	}

	tr.n -= 1
	stepResult.CellX = tr.cur_x
	stepResult.CellY = tr.cur_y
	stepResult.CellZ = tr.cur_z
	stepResult.EnterT = tr.t // TODO: check the correctness

	if tr.t_next_x < tr.t_next_y && tr.t_next_x < tr.t_next_z {
		tr.cur_x += tr.x_inc
		tr.t = tr.t_next_x
		tr.t_next_x += tr.dt_dx
	} else if tr.t_next_y < tr.t_next_x && tr.t_next_y < tr.t_next_z {
		tr.cur_y += tr.y_inc
		tr.t = tr.t_next_y
		tr.t_next_y += tr.dt_dy
	} else {
		tr.cur_z += tr.z_inc
		tr.t = tr.t_next_z
		tr.t_next_z += tr.dt_dz
	}

	stepResult.ExitT = tr.t // TODO: check the correctness
	return stepResult, true
}
