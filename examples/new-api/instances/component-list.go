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
	GenericCollider stdcomponents.GenericColliderComponentManager
	ColliderBox     stdcomponents.ColliderBoxComponentManager
	ColliderCircle  stdcomponents.ColliderCircleComponentManager
	Collision       stdcomponents.CollisionComponentManager
	SpatialIndex    stdcomponents.SpatialIndexComponentManager

	Health     components.HealthComponentManager
	Controller components.ControllerComponentManager
	PlayerTag  components.PlayerTagComponentManager
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
		GenericCollider: stdcomponents.NewGenericColliderComponentManager(),
		ColliderBox:     stdcomponents.NewColliderBoxComponentManager(),
		ColliderCircle:  stdcomponents.NewColliderCircleComponentManager(),
		Collision:       stdcomponents.NewCollisionComponentManager(),
		SpatialIndex:    stdcomponents.NewSpatialIndexComponentManager(),

		Health:     components.NewHealthComponentManager(),
		Controller: components.NewControllerComponentManager(),
		PlayerTag:  components.NewPlayerTagComponentManager(),
	}
}
