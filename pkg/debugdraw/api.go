package debugdraw

import (
	"slices"
	"time"
)

func Line(x1, y1, z1, x2, y2, z2, r, g, b, a float32, duration time.Duration) {
	if !isEnabled {
		return
	}

	*currentDrawEntries = append(*currentDrawEntries, drawEntry{
		Kind:    drawKind_Line,
		R:       r,
		G:       g,
		B:       b,
		A:       a,
		EndTime: time.Now().Add(duration),
		Values: [8]float32{
			x1, y1, z1,
			x2, y2, z2,
			0.0, 0.0,
		},
	})
}

func RectOutline(xmin, ymin, xmax, ymax, r, g, b, a float32, duration time.Duration) {
	if !isEnabled {
		return
	}

	Line(xmin, ymin, 0.0, xmax, ymin, 0.0, r, g, b, a, duration)
	Line(xmin, ymax, 0.0, xmax, ymax, 0.0, r, g, b, a, duration)
	Line(xmin, ymin, 0.0, xmin, ymax, 0.0, r, g, b, a, duration)
	Line(xmax, ymin, 0.0, xmax, ymax, 0.0, r, g, b, a, duration)
}

func IsEnabled() bool {
	return isEnabled
}

func SetEnabled(enabled bool) {
	isEnabled = enabled

	if !enabled {
		clear(nonfixedDrawEntries)
		nonfixedDrawEntries = nonfixedDrawEntries[:]
		clear(fixedDrawEntries)
		fixedDrawEntries = fixedDrawEntries[:]
	}
}

func SetFixedUpdate(fixed bool) {
	if isFixed == fixed {
		return
	}

	isFixed = fixed
	if fixed {
		// nonfixedDrawEntries = currentDrawEntries
		currentDrawEntries = &fixedDrawEntries
	} else {
		// fixedDrawEntries = currentDrawEntries
		currentDrawEntries = &nonfixedDrawEntries
	}
}

func Decay() {
	nonfixedDrawEntries = slices.DeleteFunc(nonfixedDrawEntries, func(e drawEntry) bool {
		return e.EndTime.Before(time.Now())
	})
	fixedDrawEntries = slices.DeleteFunc(fixedDrawEntries, func(e drawEntry) bool {
		return e.EndTime.Before(time.Now())
	})
}

var isEnabled = false
var isFixed = false
var nonfixedDrawEntries []drawEntry
var fixedDrawEntries []drawEntry
var currentDrawEntries = &nonfixedDrawEntries

type drawEntry struct {
	Kind       drawKind
	R, G, B, A float32
	EndTime    time.Time
	Values     [8]float32
}

type drawKind int

const (
	drawKind_Line drawKind = iota
)
