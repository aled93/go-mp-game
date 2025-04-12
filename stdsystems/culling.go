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
	s.Cameras.EachEntity(func(entity ecs.Entity) bool {
		camera := s.Cameras.Get(entity)
		cameraRect := camera.Rect()

		// Collect and sort render objects
		s.Renderables.EachEntity(func(entity ecs.Entity) bool {
			t := s.Textures.Get(entity)
			renderVisible := s.RenderVisible.Get(entity)
			aabb := s.AABBs.Get(entity)

			//TODO: rework this with future new assets manager
			if t != nil && t.Texture != nil {
				switch camera.Culling {
				case stdcomponents.Culling2DFullscreenBB:
					//TODO: textureAABB
					if renderVisible == nil {
						if aabb != nil && s.intersects(cameraRect, aabb.Rect()) {
							s.RenderVisible.Create(entity, stdcomponents.RenderVisible{})
						}
					} else {
						if !s.intersects(cameraRect, aabb.Rect()) {
							s.RenderVisible.Remove(entity)
						}
					}
				default:
					if renderVisible == nil {
						s.RenderVisible.Create(entity, stdcomponents.RenderVisible{})
					}
				}
			}

			return true
		})
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
