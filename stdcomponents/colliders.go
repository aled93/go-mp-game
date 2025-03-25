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

type BoxCollider struct {
	WH     vectors.Vec2
	Offset vectors.Vec2
	Layer  CollisionLayer
	Mask   CollisionMask
}

func (c *BoxCollider) GetSupport(direction vectors.Vec2, transform *Transform2d) vectors.Vec2 {
	var maxDistance float32 = -math.MaxFloat32
	var maxPoint vectors.Vec2

	vertices := [4]vectors.Vec2{
		{X: c.WH.X, Y: c.WH.Y},
		{X: 0, Y: c.WH.Y},
		{X: 0, Y: 0},
		{X: c.WH.X, Y: 0},
	}

	for i := range vertices {
		vertex := &vertices[i]
		worldVertex := vertex.Sub(c.Offset).Rotate(transform.Rotation)

		distance := worldVertex.Dot(direction)
		if distance > maxDistance {
			maxDistance = distance
			maxPoint = worldVertex
		}
	}

	return maxPoint.Mul(transform.Scale).Add(transform.Position)
}

type BoxColliderComponentManager = ecs.ComponentManager[BoxCollider]

func NewBoxColliderComponentManager() BoxColliderComponentManager {
	return ecs.NewComponentManager[BoxCollider](ColliderBoxComponentId)
}

type CircleCollider struct {
	Radius float32
	Layer  CollisionLayer
	Mask   CollisionMask
	Offset vectors.Vec2
}

func (c *CircleCollider) GetSupport(direction vectors.Vec2, transform *Transform2d) vectors.Vec2 {
	if direction.LengthSquared() == 0 {
		return transform.Position
	}
	radius := c.Radius * transform.Scale.X
	dirNorm := direction.Normalize()
	return transform.Position.Add(dirNorm.Scale(radius))
}

type CircleColliderComponentManager = ecs.ComponentManager[CircleCollider]

func NewCircleColliderComponentManager() CircleColliderComponentManager {
	return ecs.NewComponentManager[CircleCollider](ColliderCircleComponentId)
}

type PolygonCollider struct {
	Vertices []vectors.Vec2
	Layer    CollisionLayer
	Mask     CollisionMask
	Offset   vectors.Vec2
}

func (c *PolygonCollider) GetSupport(direction vectors.Vec2, transform *Transform2d) vectors.Vec2 {
	maxDot := math.Inf(-1)
	var maxVertex vectors.Vec2

	for _, v := range c.Vertices {
		scaled := v.Mul(transform.Scale)
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

type GenericCollider struct {
	Shape  ColliderShape
	Layer  CollisionLayer
	Mask   CollisionMask
	Offset vectors.Vec2
}

type GenericColliderComponentManager = ecs.ComponentManager[GenericCollider]

func NewGenericColliderComponentManager() GenericColliderComponentManager {
	return ecs.NewComponentManager[GenericCollider](GenericColliderComponentId)
}
