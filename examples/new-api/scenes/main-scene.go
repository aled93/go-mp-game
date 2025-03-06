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

package scenes

import (
	"gomp"
	"gomp/examples/new-api/instances"
	"gomp/pkg/ecs"
	"time"
)

func NewMainScene() MainScene {
	return MainScene{
		World: ecs.NewWorld(instances.NewComponentList(), instances.NewSystemList()),
	}
}

type MainScene struct {
	Game  *gomp.Game
	World instances.World
}

func (s *MainScene) Id() gomp.SceneId {
	return MainSceneId
}

func (s *MainScene) Init() {
	s.World.Init()

	// Network receive
	s.World.Systems.Network.Init()
	s.World.Systems.NetworkReceive.Init()

	// Scenes
	s.World.Systems.Player.Init()
	s.World.Systems.Velocity.Init()

	s.World.Systems.Collision.Init()

	// Network patches
	s.World.Systems.NetworkSend.Init()

	// Animation
	s.World.Systems.AnimationSpriteMatrix.Init()
	s.World.Systems.AnimationPlayer.Init()

	s.World.Systems.SpriteMatrix.Init()
	s.World.Systems.YSort.Init()

	// Render
	s.World.Systems.Render.Init()
	s.World.Systems.Debug.Init()
	s.World.Systems.AssetLib.Init()
}

func (s *MainScene) Update(dt time.Duration) gomp.SceneId {
	// Network receive
	s.World.Systems.NetworkReceive.Run(dt)
	s.World.Systems.Player.Run()

	return MainSceneId
}

func (s *MainScene) FixedUpdate(dt time.Duration) {
	// Network send
	s.World.Systems.Network.Run(dt)

	s.World.Systems.Velocity.Run(dt)
	s.World.Systems.SpatialCollision.Run(dt)
	s.World.Systems.CollisionHandler.Run(dt)
	s.World.Systems.NetworkSend.Run(dt)
}

func (s *MainScene) Render(dt time.Duration) {
	// Animation
	s.World.Systems.AnimationSpriteMatrix.Run()
	s.World.Systems.AnimationPlayer.Run()

	s.World.Systems.SpriteMatrix.Run()
	s.World.Systems.Debug.Run()
	s.World.Systems.AssetLib.Run()
	s.World.Systems.YSort.Run()

	shouldContinue := s.World.Systems.Render.Run(dt)
	if !shouldContinue {
		s.Game.SetShouldDestroy(true)
		return
	}
}

func (s *MainScene) Destroy() {
	s.World.Destroy()
	// Network intents
	s.World.Systems.Network.Destroy()
	s.World.Systems.NetworkReceive.Destroy()

	s.World.Systems.Player.Destroy()

	s.World.Systems.Collision.Destroy()

	// Network patches
	s.World.Systems.NetworkSend.Destroy()

	// Animation
	s.World.Systems.AnimationSpriteMatrix.Destroy()
	s.World.Systems.AnimationPlayer.Destroy()

	s.World.Systems.SpriteMatrix.Destroy()
	s.World.Systems.YSort.Destroy()

	// Render
	s.World.Systems.Debug.Destroy()
	s.World.Systems.AssetLib.Destroy()
	s.World.Systems.Render.Destroy()
}

func (s *MainScene) OnEnter() {

}

func (s *MainScene) OnExit() {

}

var _ gomp.AnyScene = (*MainScene)(nil)
