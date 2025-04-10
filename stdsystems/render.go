/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- HromRu Donated 1 500 RUB

Thank you for your support!
*/

package stdsystems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"slices"
	"time"
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct {
	EntityManager                      *ecs.EntityManager
	RlTexturePros                      *stdcomponents.RLTextureProComponentManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	AnimationPlayers                   *stdcomponents.AnimationPlayerComponentManager
	Tints                              *stdcomponents.TintComponentManager
	Flips                              *stdcomponents.FlipComponentManager
	Renderables                        *stdcomponents.RenderableComponentManager
	AnimationStates                    *stdcomponents.AnimationStateComponentManager
	Sprites                            *stdcomponents.SpriteComponentManager
	SpriteMatrixes                     *stdcomponents.SpriteMatrixComponentManager
	RenderOrders                       *stdcomponents.RenderOrderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	AABBs                              *stdcomponents.AABBComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTrees                           *stdcomponents.BvhTreeComponentManager
	CamerasManager                     *stdcomponents.CameraComponentManager
	FrameBuffer2D                      *stdcomponents.FrameBuffer2DComponentManager

	renderTextures []rl.RenderTexture2D
	frames         []stdcomponents.FrameBuffer2D

	monitorWidth  int
	monitorHeight int

	debug bool
}

func (s *RenderSystem) Init() {
	s.monitorWidth = rl.GetScreenWidth()
	s.monitorHeight = rl.GetScreenHeight()
	s.renderTextures = make([]rl.RenderTexture2D, 0, s.CamerasManager.Len())
	s.frames = make([]stdcomponents.FrameBuffer2D, 0, s.FrameBuffer2D.Len())
}

func (s *RenderSystem) Run(dt time.Duration) bool {

	s.FrameBuffer2D.EachComponent(func(c *stdcomponents.FrameBuffer2D) bool {
		s.frames = append(s.frames, *c)
		return true
	})
	slices.SortFunc(s.frames, func(a, b stdcomponents.FrameBuffer2D) int {
		return int(a.Layer - b.Layer)
	})

	rl.BeginDrawing()

	for _, texture := range s.frames {
		rl.BeginBlendMode(texture.BlendMode)
		rl.DrawTexturePro(texture.Texture.Texture,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(texture.Texture.Texture.Width),
				Height: -float32(texture.Texture.Texture.Height),
			},
			texture.Dst,
			rl.Vector2{},
			texture.Rotation,
			texture.Tint,
		)
		rl.EndBlendMode()
	}

	rl.EndDrawing()

	s.frames = s.frames[:0]
	return false
}

func (s *RenderSystem) Destroy() {
}

type RenderInjector struct {
	EntityManager                      *ecs.EntityManager
	RlTexturePros                      *stdcomponents.RLTextureProComponentManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	AnimationPlayers                   *stdcomponents.AnimationPlayerComponentManager
	Tints                              *stdcomponents.TintComponentManager
	Flips                              *stdcomponents.FlipComponentManager
	Renderables                        *stdcomponents.RenderableComponentManager
	AnimationStates                    *stdcomponents.AnimationStateComponentManager
	Sprites                            *stdcomponents.SpriteComponentManager
	SpriteMatrixes                     *stdcomponents.SpriteMatrixComponentManager
	RenderOrders                       *stdcomponents.RenderOrderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	AABBs                              *stdcomponents.AABBComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTrees                           *stdcomponents.BvhTreeComponentManager
	Camera                             *stdcomponents.CameraComponentManager
	RenderTexture2D                    *stdcomponents.FrameBuffer2DComponentManager
}

func (s *RenderSystem) InjectWorld(injector *RenderInjector) {
	s.EntityManager = injector.EntityManager
	s.RlTexturePros = injector.RlTexturePros
	s.Positions = injector.Positions
	s.Rotations = injector.Rotations
	s.Scales = injector.Scales
	s.AnimationPlayers = injector.AnimationPlayers
	s.Tints = injector.Tints
	s.Flips = injector.Flips
	s.Renderables = injector.Renderables
	s.AnimationStates = injector.AnimationStates
	s.Sprites = injector.Sprites
	s.SpriteMatrixes = injector.SpriteMatrixes
	s.RenderOrders = injector.RenderOrders
	s.BoxColliders = injector.BoxColliders
	s.CircleColliders = injector.CircleColliders
	s.AABBs = injector.AABBs
	s.Collisions = injector.Collisions
	s.ColliderSleepStateComponentManager = injector.ColliderSleepStateComponentManager
	s.BvhTrees = injector.BvhTrees
	s.CamerasManager = injector.Camera
	s.FrameBuffer2D = injector.RenderTexture2D
}
