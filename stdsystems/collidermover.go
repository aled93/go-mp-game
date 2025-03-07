package stdsystems

import (
	"gomp/pkg/collision"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

type ColliderMoverSystem struct {
	Colliders   *stdcomponents.ColliderComponentManager
	Velocities  *stdcomponents.VelocityComponentManager
	Positions   *stdcomponents.PositionComponentManager
	PhysObjects *stdcomponents.PhysObjectComponentManager
	PhysSpaces  *stdcomponents.PhysSpaceComponentManager
}

func (sys *ColliderMoverSystem) Init() {
	//
}

func (sys *ColliderMoverSystem) Run(dt time.Duration) {
	sys.Colliders.Each(func(ent ecs.Entity, collider *stdcomponents.Collider) bool {
		vel := sys.Velocities.Get(ent)
		if vel == nil {
			return true
		}

		pos := sys.Positions.Get(ent)
		if pos == nil {
			return true
		}

		physSpace := sys.PhysSpaces.Get(ent)
		if physSpace == nil || physSpace.Space2D == nil {
			return true
		}

		physObj := sys.PhysObjects.Get(ent)
		if physObj == nil {
			physobj := physSpace.CreateObject(&collision.ObjectCreateParams{
				Shape:         collision.Box,
				ShapeSize:     128.0,
				CollisionMask: collision.CollideAll,
				X:             float64(pos.X),
				Y:             float64(pos.Y),
			})
			physObj = sys.PhysObjects.Create(ent, stdcomponents.PhysObject{Object: physobj})
		}

		if vel.X == 0.0 && vel.Y == 0.0 {
			return true
		}

		physObj.Move(float64(vel.X)*dt.Seconds(), float64(vel.Y)*dt.Seconds())

		newX, newY := physObj.Pos()
		pos.X = float32(newX)
		pos.Y = float32(newY)

		return true
	})
}

func (sys *ColliderMoverSystem) Destroy() {
	//
}
