/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdcomponents

import (
	"gomp/pkg/ecs"
)

// StdComponentIds MUST always be the last
const (
	InvalidComponentId ecs.ComponentId = iota
	TransformComponentId
	PositionComponentId
	RotationComponentId
	ScaleComponentId
	FlipComponentId
	VelocityComponentId
	SpriteComponentId
	SpriteSheetComponentId
	SpriteMatrixComponentId
	RLTextureProComponentId
	AnimationPlayerComponentId
	AnimationStateComponentId
	TintComponentId
	NetworkComponentId
	RenderableComponentId
	YSortComponentId
	RenderOrderComponentId
	GenericColliderComponentId
	ColliderBoxComponentId
	ColliderCircleComponentId
	ColliderSleepStateComponentId
	PolygonColliderComponentId
	CollisionComponentId
	SpatialIndexComponentId
	AABBComponentId
	RigidBodyComponentId
	BvhTreeComponentId
	CollisionGridComponentId

	StdComponentIds // StdComponentIds MUST always be the last
)
