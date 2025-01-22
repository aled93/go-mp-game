// Package purely to resolve cyclic import
package gravity

import "gomp_game/pkgs/spatial"

type QuadNodeUserData struct {
	GX, GY float64
	Mass   float64
}

var QTree *spatial.QuadTree2D[QuadNodeUserData, any]
