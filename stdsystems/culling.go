/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"time"
)

func NewCullingSystem() CullingSystem {
	return CullingSystem{}
}

type CullingSystem struct {
	Renderables     *stdcomponents.RenderableComponentManager
	RenderVisible   *stdcomponents.RenderVisibleComponentManager
	RenderOrders    *stdcomponents.RenderOrderComponentManager
	Textures        *stdcomponents.RLTextureProComponentManager
	AABBs           *stdcomponents.AABBComponentManager
	Cameras         *stdcomponents.CameraComponentManager
	RenderTexture2D *stdcomponents.FrameBuffer2DComponentManager
}

func (s *CullingSystem) Init() {
}

func (s *CullingSystem) Run(dt time.Duration) {
	s.Renderables.EachComponentParallel(2048, func(r *stdcomponents.Renderable, i int) bool {
		r.Observed = false
		return true
	})
	s.Cameras.EachEntity(func(entity ecs.Entity) bool {
		camera := s.Cameras.Get(entity)
		cameraRect := camera.Rect()
		s.Renderables.EachEntity(func(entity ecs.Entity) bool {
			renderable := s.Renderables.Get(entity)
			//renderVisible := s.RenderVisible.Get(entity)
			aabb := s.AABBs.Get(entity)

			switch camera.Culling {
			case stdcomponents.Culling2DFullscreenBB:
				//TODO: textureAABB
				if aabb == nil {
					renderable.Observed = true
					return true
				}
				if s.intersects(cameraRect, aabb.Rect()) {
					renderable.Observed = true
				}

			default:
				renderable.Observed = true
			}

			return true
		})
		return true
	})
	s.Renderables.EachEntity(func(entity ecs.Entity) bool {
		renderable := s.Renderables.Get(entity)
		visible := s.RenderVisible.Get(entity)
		if visible == nil {
			if renderable.Observed {
				s.RenderVisible.Create(entity, stdcomponents.RenderVisible{})
			}
		} else {
			if !renderable.Observed {
				s.RenderVisible.Remove(entity)
			}
		}
		return true
	})
}

func (_ *CullingSystem) intersects(rect1, rect2 vectors.Rectangle) bool {
	return rect1.X < rect2.X+rect2.Width &&
		rect1.X+rect1.Width > rect2.X &&
		rect1.Y < rect2.Y+rect2.Height &&
		rect1.Y+rect1.Height > rect2.Y
}

func (s *CullingSystem) Destroy() {
}
