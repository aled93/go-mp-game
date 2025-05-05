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
	"math"
)

type ColliderShape uint8

const (
	InvalidColliderShape ColliderShape = iota
	BoxColliderShape
	CircleColliderShape
	PolygonColliderShape
)

type CollisionMask uint64

func (m CollisionMask) HasLayer(layer CollisionLayer) bool {
	return m&(1<<layer) != 0
}

type CollisionLayer = CollisionMask

const (
	CollisionLayerNone CollisionLayer = 0
)

// ===========================
// Box Collider
// ===========================

type BoxCollider struct {
	WH         util.Vec2
	Offset     util.Vec2
	Layer      CollisionLayer
	Mask       CollisionMask
	AllowSleep bool
}

func (c *BoxCollider) GetSupport(direction util.Vec2, transform Transform2d) util.Vec2 {
	// Precompute rotation terms once
	cos := float32(math.Cos(transform.Rotation))
	sin := float32(math.Sin(transform.Rotation))

	// Inverse-rotate direction to local space (avoids per-vertex rotation)
	localDir := util.NewVec2(
		direction.X*cos+direction.Y*sin,
		-direction.X*sin+direction.Y*cos,
	)

	// Branchless selection using sign bits (Go-optimized)
	xSign := math.Float32bits(localDir.X) >> 31
	ySign := math.Float32bits(localDir.Y) >> 31
	localSupport := util.NewVec2(
		c.WH.X*(1-float32(xSign)), // 0 if negative, WH.X otherwise
		c.WH.Y*(1-float32(ySign)),
	)

	// Apply offset and rotate to world space
	vertex := localSupport.Subtract(c.Offset)
	rotated := util.NewVec2(
		vertex.X*cos-vertex.Y*sin,
		vertex.X*sin+vertex.Y*cos,
	)

	return rotated.Scale(transform.Scale).Add(transform.Position)
}

type BoxColliderComponentManager = ecs.ComponentManager[BoxCollider]

func NewBoxColliderComponentManager() BoxColliderComponentManager {
	return ecs.NewComponentManager[BoxCollider](ColliderBoxComponentId)
}

// ===========================
// Circle Collider
// ===========================

type CircleCollider struct {
	Radius     float32
	Layer      CollisionLayer
	Mask       CollisionMask
	Offset     util.Vec2
	AllowSleep bool
}

func (c *CircleCollider) GetSupport(direction util.Vec2, transform Transform2d) util.Vec2 {
	var radiusWithOffset util.Vec2
	// Handle zero direction to avoid division by zero
	if direction.X == 0 && direction.Y == 0 {
		// Fallback to a default direction (e.g., right)
		defaultDir := util.Vec2{X: c.Radius, Y: 0}
		radiusWithOffset = defaultDir.Subtract(c.Offset).Scale(transform.Scale)
	} else {
		// Compute scaled direction without trigonometry
		mag := float32(math.Hypot(float64(direction.X), float64(direction.Y)))
		invMag := c.Radius / mag
		scaledDir := util.Vec2{
			X: direction.X * invMag,
			Y: direction.Y * invMag,
		}
		// Apply offset, scale, and translation
		radiusWithOffset = scaledDir.Subtract(c.Offset).Scale(transform.Scale)
	}

	return transform.Position.Add(radiusWithOffset)
}

type CircleColliderComponentManager = ecs.ComponentManager[CircleCollider]

func NewCircleColliderComponentManager() CircleColliderComponentManager {
	return ecs.NewComponentManager[CircleCollider](ColliderCircleComponentId)
}

// ===========================
// Polygon Collider
// ===========================

type PolygonCollider struct {
	Vertices   []util.Vec2
	Layer      CollisionLayer
	Mask       CollisionMask
	Offset     util.Vec2
	AllowSleep bool
}

func (c *PolygonCollider) GetSupport(direction util.Vec2, transform Transform2d) util.Vec2 {
	maxDot := math.Inf(-1)
	var maxVertex util.Vec2

	for _, v := range c.Vertices {
		scaled := v.Scale(transform.Scale)
		rotated := scaled.Rotate(transform.Rotation)
		worldVertex := transform.Position.Add(rotated)
		dot := float64(worldVertex.Dot(direction))
		if dot > maxDot {
			maxDot = dot
			maxVertex = worldVertex
		}
	}
	return maxVertex
}

type PolygonColliderComponentManager = ecs.ComponentManager[PolygonCollider]

func NewPolygonColliderComponentManager() PolygonColliderComponentManager {
	return ecs.NewComponentManager[PolygonCollider](PolygonColliderComponentId)
}

// ===========================
// Generic Collider
// ===========================

type GenericCollider struct {
	Shape      ColliderShape
	Layer      CollisionLayer
	Mask       CollisionMask
	Offset     util.Vec2
	AllowSleep bool
}

type GenericColliderComponentManager = ecs.ComponentManager[GenericCollider]

func NewGenericColliderComponentManager() GenericColliderComponentManager {
	return ecs.NewComponentManager[GenericCollider](GenericColliderComponentId)
}

type ColliderSleepState struct{}

type ColliderSleepStateComponentManager = ecs.ComponentManager[ColliderSleepState]

func NewColliderSleepStateComponentManager() ColliderSleepStateComponentManager {
	return ecs.NewComponentManager[ColliderSleepState](ColliderSleepStateComponentId)
}
