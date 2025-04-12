/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"cmp"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"slices"
	"time"
)

func NewRender2DCamerasSystem() Render2DCamerasSystem {
	return Render2DCamerasSystem{}
}

type Render2DCamerasSystem struct {
	Renderables     *stdcomponents.RenderableComponentManager
	RenderOrders    *stdcomponents.RenderOrderComponentManager
	Textures        *stdcomponents.RLTextureProComponentManager
	AABBs           *stdcomponents.AABBComponentManager
	Cameras         *stdcomponents.CameraComponentManager
	RenderTexture2D *stdcomponents.FrameBuffer2DComponentManager
	renderObjects   []renderObject
}

type renderObject struct {
	texture    *stdcomponents.RLTexturePro
	renderable *stdcomponents.Renderable
	aabb       *stdcomponents.AABB
	order      float32
}

func (s *Render2DCamerasSystem) Init() {
	s.renderObjects = make([]renderObject, 0, s.Renderables.Len())
}

func (s *Render2DCamerasSystem) Run(dt time.Duration) {
	s.Cameras.EachEntity(func(entity ecs.Entity) bool {
		camera := s.Cameras.Get(entity)
		cameraRect := camera.Rect()
		renderTexture := s.RenderTexture2D.Get(entity)

		// Collect and sort render objects
		s.Renderables.EachEntity(func(entity ecs.Entity) bool {
			r := s.Renderables.Get(entity)
			t := s.Textures.Get(entity)
			o := s.RenderOrders.Get(entity)
			aabb := s.AABBs.Get(entity)

			if t != nil {
				switch camera.Culling {
				case stdcomponents.Culling2DFullscreenBB:
					if aabb != nil && intersects(cameraRect, aabb.Rect()) {
						s.renderObjects = append(s.renderObjects, renderObject{
							texture:    t,
							renderable: r,
							order:      o.CalculatedZ,
							aabb:       aabb,
						})
					}
				default:
					s.renderObjects = append(s.renderObjects, renderObject{
						texture:    t,
						renderable: r,
						order:      o.CalculatedZ,
						aabb:       aabb,
					})
				}
			}

			return true
		})

		slices.SortFunc(s.renderObjects, func(a, b renderObject) int {
			return cmp.Compare(a.order, b.order)
		})

		// Draw render objects
		rl.BeginTextureMode(renderTexture.Texture)
		rl.BeginMode2D(camera.Camera2D)
		rl.ClearBackground(camera.BGColor)

		for _, obj := range s.renderObjects {
			if camera.Layer&obj.renderable.CameraMask != 0 {
				assert.Nil(obj.texture, "EntityTexturePro is nil")
				rl.DrawTexturePro(*obj.texture.Texture, obj.texture.Frame, obj.texture.Dest, obj.texture.Origin, obj.texture.Rotation, obj.texture.Tint)
			}
		}

		rl.EndMode2D()
		rl.EndTextureMode()

		s.renderObjects = s.renderObjects[:0]
		return true
	})
}

func intersects(rect1, rect2 vectors.Rectangle) bool {
	return rect1.X < rect2.X+rect2.Width &&
		rect1.X+rect1.Width > rect2.X &&
		rect1.Y < rect2.Y+rect2.Height &&
		rect1.Y+rect1.Height > rect2.Y
}

func (s *Render2DCamerasSystem) Destroy() {
}
