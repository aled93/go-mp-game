/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/examples/raylib-ecs/components"
	ecs2 "gomp/pkg/ecs"
)

// TextureRenderScale is a system that sets Scale of textureRender
type trMirroredController struct{}

func (s *trMirroredController) Init(world *ecs2.EntityManager)        {}
func (s *trMirroredController) FixedUpdate(world *ecs2.EntityManager) {}
func (s *trMirroredController) Update(world *ecs2.EntityManager) {
	// Get component managers
	mirroreds := components.MirroredService.GetManager(world)
	textureRenders := components.TextureRenderService.GetManager(world)

	// Update sprites and spriteRenders
	textureRenders.AllParallel(func(entity ecs2.Entity, tr *components.TextureRender) bool {
		if tr == nil {
			return true
		}

		mirrored := mirroreds.Get(entity)
		if mirrored == nil {
			return true
		}

		if mirrored.X {
			tr.Frame.Width *= -1
		}
		if mirrored.Y {
			tr.Frame.Height *= -1
		}

		return true
	})
}
func (s *trMirroredController) Destroy(world *ecs2.EntityManager) {}
