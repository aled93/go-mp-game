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

package instances

import (
	"gomp/examples/new-api/components"
	"gomp/stdcomponents"
)

type ComponentList struct {
	Position        stdcomponents.PositionComponentManager
	Rotation        stdcomponents.RotationComponentManager
	Scale           stdcomponents.ScaleComponentManager
	Velocity        stdcomponents.VelocityComponentManager
	Flip            stdcomponents.FlipComponentManager
	SpriteMatrix    stdcomponents.SpriteMatrixComponentManager
	Tint            stdcomponents.TintComponentManager
	AnimationPlayer stdcomponents.AnimationPlayerComponentManager
	AnimationState  stdcomponents.AnimationStateComponentManager
	RLTexturePro    stdcomponents.RLTextureProComponentManager
	Network         stdcomponents.NetworkComponentManager
	Renderable      stdcomponents.RenderableComponentManager
	YSort           stdcomponents.YSortComponentManager
	RenderOrder     stdcomponents.RenderOrderComponentManager
	PhysSpace       stdcomponents.PhysSpaceComponentManager
	PhysObject      stdcomponents.PhysObjectComponentManager

	Health     components.HealthComponentManager
	Controller components.ControllerComponentManager
}

func NewComponentList() ComponentList {
	return ComponentList{
		Position:        stdcomponents.NewPositionComponentManager(),
		Rotation:        stdcomponents.NewRotationComponentManager(),
		Scale:           stdcomponents.NewScaleComponentManager(),
		Velocity:        stdcomponents.NewVelocityComponentManager(),
		Flip:            stdcomponents.NewFlipComponentManager(),
		SpriteMatrix:    stdcomponents.NewSpriteMatrixComponentManager(),
		Tint:            stdcomponents.NewTintComponentManager(),
		AnimationPlayer: stdcomponents.NewAnimationPlayerComponentManager(),
		AnimationState:  stdcomponents.NewAnimationStateComponentManager(),
		RLTexturePro:    stdcomponents.NewRlTextureProComponentManager(),
		Network:         stdcomponents.NewNetworkComponentManager(),
		Renderable:      stdcomponents.NewRenderableComponentManager(),
		YSort:           stdcomponents.NewYSortComponentManager(),
		RenderOrder:     stdcomponents.NewRenderOrderComponentManager(),
		PhysSpace:       stdcomponents.NewPhysSpaceComponentManager(),
		PhysObject:      stdcomponents.NewPhysObjectComponentManager(),

		Health:     components.NewHealthComponentManager(),
		Controller: components.NewControllerComponentManager(),
	}
}
