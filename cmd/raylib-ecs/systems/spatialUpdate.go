/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
	"gomp_game/pkgs/spatial"
)

type spatialUpdateController struct {
	positions  *ecs.ComponentManager[components.Position]
	spatialIds *ecs.ComponentManager[components.SpatialEnt]
}

func (s *spatialUpdateController) Init(world *ecs.World) {
	gravity.QTree = spatial.NewQuadTree2D[gravity.QuadNodeUserData, any](800.0, 600.0, 16, 64)
	s.positions = components.PositionService.GetManager(world)
	s.spatialIds = components.SpatialEntService.GetManager(world)
}

func (s *spatialUpdateController) Update(world *ecs.World) {
	s.positions.AllParallel(func(entId ecs.EntityID, pos *components.Position) bool {
		spatId := s.spatialIds.Get(entId)
		if spatId == nil {
			if ent, ok := gravity.QTree.AddPoint(float64(pos.X), float64(pos.Y)); ok {
				s.spatialIds.Create(entId, components.SpatialEnt{
					Ent: ent,
				})
			}

			return true
		}

		gravity.QTree.UpdatePosition(spatId.Ent, float64(pos.X), float64(pos.Y))

		return true
	})

	gravity.QTree.Maintain()
}

func (s *spatialUpdateController) FixedUpdate(world *ecs.World) {}
func (s *spatialUpdateController) Destroy(world *ecs.World)     {}
