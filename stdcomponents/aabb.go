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
	"gomp/pkg/util"
)

type AABB struct {
	Min util.Vec2
	Max util.Vec2
}

func (a AABB) Center() util.Vec2 {
	return a.Min.Add(a.Max).ScaleScalar(0.5)
}

func (a AABB) Rect() util.Rect {
	return util.NewRectFromMinMax(a.Min, a.Max)
}

type AABBComponentManager = ecs.ComponentManager[AABB]

func NewAABBComponentManager() AABBComponentManager {
	return ecs.NewComponentManager[AABB](AABBComponentId)
}
