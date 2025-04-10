/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package stdcomponents

import (
	"gomp/pkg/ecs"
	"gomp/vectors"
)

type SpatialIndex struct {
	X, Y int
}

func (i SpatialIndex) ToVec2() vectors.Vec2 {
	return vectors.Vec2{X: float32(i.X), Y: float32(i.Y)}
}

type SpatialIndexComponentManager = ecs.ComponentManager[SpatialIndex]

func NewSpatialIndexComponentManager() SpatialIndexComponentManager {
	return ecs.NewComponentManager[SpatialIndex](SpatialIndexComponentId)
}
