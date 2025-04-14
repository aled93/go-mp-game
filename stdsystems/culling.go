/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"runtime"
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
	numWorkers      int
}

func (s *CullingSystem) Init() {
	s.numWorkers = runtime.NumCPU() - 2
}

func (s *CullingSystem) Run(dt time.Duration) {
	s.Renderables.EachComponentParallel(s.numWorkers)(func(r *stdcomponents.Renderable, i int) bool {
		r.Observed = false
		return true
	})

	s.Cameras.EachEntity()(func(entity ecs.Entity) bool {
		camera := s.Cameras.GetUnsafe(entity)
		cameraRect := camera.Rect()

		s.Renderables.EachEntityParallel(s.numWorkers)(func(entity ecs.Entity, _ int) bool {
			renderable := s.Renderables.GetUnsafe(entity)
			assert.NotNil(renderable)

			texture := s.Textures.GetUnsafe(entity)
			assert.NotNil(texture)

			textureRect := texture.Rect()

			if s.intersects(cameraRect, textureRect) {
				renderable.Observed = true
			}
			return true
		})
		return true
	})

	var accRenderVisibleCreate = make([][]ecs.Entity, s.numWorkers)
	var accRenderVisibleDelete = make([][]ecs.Entity, s.numWorkers)
	s.Renderables.EachEntityParallel(s.numWorkers)(func(entity ecs.Entity, workerId int) bool {
		renderable := s.Renderables.GetUnsafe(entity)
		assert.NotNil(renderable)
		if !s.RenderVisible.Has(entity) {
			if renderable.Observed {
				accRenderVisibleCreate[workerId] = append(accRenderVisibleCreate[workerId], entity)
			}
		} else {
			if !renderable.Observed {
				accRenderVisibleDelete[workerId] = append(accRenderVisibleDelete[workerId], entity)
			}
		}
		return true
	})
	for a := range accRenderVisibleCreate {
		for _, entity := range accRenderVisibleCreate[a] {
			s.RenderVisible.Create(entity, stdcomponents.RenderVisible{})
		}
	}
	for a := range accRenderVisibleDelete {
		for _, entity := range accRenderVisibleDelete[a] {
			s.RenderVisible.Delete(entity)
		}
	}
}

func (_ *CullingSystem) intersects(rect1, rect2 vectors.Rectangle) bool {
	return rect1.X < rect2.X+rect2.Width &&
		rect1.X+rect1.Width > rect2.X &&
		rect1.Y < rect2.Y+rect2.Height &&
		rect1.Y+rect1.Height > rect2.Y
}

func (s *CullingSystem) Destroy() {
}
