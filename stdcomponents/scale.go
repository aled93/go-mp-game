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

import "gomp/pkg/ecs"

type Scale struct {
	X, Y float32
}

type ScaleComponentManager = ecs.ComponentManager[Scale]

func NewScaleComponentManager(world *ecs.World) *ecs.ComponentManager[Scale] {
	return ecs.NewComponentManager[Scale](world, SCALE_ID)
}
