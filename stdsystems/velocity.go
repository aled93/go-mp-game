/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/collision"
	"gomp/pkg/debugdraw"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"time"
)

func NewVelocitySystem() VelocitySystem {
	return VelocitySystem{}
}

type VelocitySystem struct {
	Velocities  *stdcomponents.VelocityComponentManager
	Positions   *stdcomponents.PositionComponentManager
	PhysSpaces  *stdcomponents.PhysSpaceComponentManager
	PhysObjects *stdcomponents.PhysObjectComponentManager
}

func (s *VelocitySystem) Init() {}
func (s *VelocitySystem) Run(dt time.Duration) {
	dtSec := float32(dt.Seconds())
	s.Velocities.EachEntity(func(e ecs.Entity) bool {
		velocity := s.Velocities.Get(e)
		position := s.Positions.Get(e)
		physSpace := s.PhysSpaces.Get(e)
		physObj := s.PhysObjects.Get(e)

		if physSpace == nil {
			position.X += velocity.X * dtSec
			position.Y += velocity.Y * dtSec
		} else {
			if physObj == nil {
				physobj := physSpace.CreateObject(&collision.ObjectCreateParams{
					Shape:         collision.Box,
					ShapeSize:     32.0,
					CollisionMask: collision.CollideAll,
					X:             float64(position.X),
					Y:             float64(position.Y),
				})
				physObj = s.PhysObjects.Create(e, stdcomponents.PhysObject{Object: physobj})
			}

			x, y := physObj.Pos()
			sz := physObj.Size()
			debugdraw.RectOutline(float32(x-sz), float32(y-sz), float32(x+sz), float32(y+sz), 0.0, 1.0, 0.0, 1.0, 0)

			if velocity.X == 0.0 && velocity.Y == 0.0 {
				return true
			}

			physObj.Move(float64(velocity.X)*dt.Seconds(), float64(velocity.Y)*dt.Seconds())

			newX, newY := physObj.Pos()
			position.X = float32(newX)
			position.Y = float32(newY)
		}

		return true
	})
}
func (s *VelocitySystem) Destroy() {}
