/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- HromRu Donated 2 500 RUB
<- kotyamatroskin Donated 400 RUB
<- r_uslashk_a Donated 100 RUB
<- mitwelve Donated 100 RUB

Thank you for your support!
*/

package scenes

import (
	"gomp"
	"gomp/examples/new-api/instances"
	"gomp/pkg/ecs"
	"time"
)

/*
TITLE:
- Roddsteroids
- Asteroidds
- Asteroids
+ Assterodds

USER FLOW:
- press space to start
+ spaceship AD - turn ; WS - gas
+ spaceship HP = 3
+ asteroids coming from the top
+ asteroids HP=3-9
- random bonus from asteroid (hp, weapon, move speed)
- asteroid spawn rate increases over the time
- player scores (optional save scores over the session)

ENTITIES and COMPONENTS:
- scene-manager: sceneState, playerScore, playerHighScore
+ spaceship: position, rotation, scale, hp, velocity, sprite, collider, playerTag, weapon
+ asteroid: position, rotation, scale, hp, velocity, sprite, collider, asteroidTag, scoreReward
+ asteroid-spawner: position, velocity, asteroidSpawnerTag
+ bullet: position, rotation, scale, velocity, sprite, collider, damage, bulletTag
- bonus: position, rotation, scale, velocity, sprite, collider, bonusTag

COMPONENTS:
- position, rotation, scale, velocity
- hp, damage, scoreReward, weapon
- asteroidTag, bulletTag, bonusTag, playerTag, asteroidSpawnerTag
- sprite, collider
*/

func NewAssteroddScene() AssteroddScene {
	return AssteroddScene{
		World: ecs.NewWorld(instances.NewComponentList(), instances.NewSystemList()),
	}
}

type AssteroddScene struct {
	Game  *gomp.Game
	World instances.World
}

func (s *AssteroddScene) Id() gomp.SceneId {
	return AssteroddSceneId
}

func (s *AssteroddScene) Init() {
	s.World.Init()
	s.World.Systems.ColliderSystem.Init()

	// Scenes
	s.World.Systems.SpaceSpawner.Init()
	s.World.Systems.AssteroddSystem.Init()
	s.World.Systems.SpaceshipIntents.Init()
	s.World.Systems.Velocity.Init()

	s.World.Systems.CollisionDetectionBVH.Init()
	s.World.Systems.CollisionResolution.Init()

	// Animation
	s.World.Systems.AnimationSpriteMatrix.Init()
	s.World.Systems.AnimationPlayer.Init()

	s.World.Systems.SpriteMatrix.Init()
	s.World.Systems.Sprite.Init()
	s.World.Systems.YSort.Init()

	// RenderAssterodd
	s.World.Systems.RenderAssterodd.Init()
	s.World.Systems.Debug.Init()
	s.World.Systems.AssetLib.Init()
	s.World.Systems.Audio.Init()
}

func (s *AssteroddScene) Update(dt time.Duration) gomp.SceneId {
	s.World.Systems.ColliderSystem.Run(dt)
	s.World.Systems.AssteroddSystem.Run(dt)

	return AssteroddSceneId
}

func (s *AssteroddScene) FixedUpdate(dt time.Duration) {
	s.World.Systems.SpaceshipIntents.Run(dt)
	s.World.Systems.Velocity.Run(dt)
	s.World.Systems.SpaceSpawner.Run(dt)
	s.World.Systems.CollisionDetectionBVH.Run(dt)
	s.World.Systems.CollisionResolution.Run(dt)
	s.World.Systems.CollisionHandler.Run(dt)
	s.World.Systems.Hp.Run(dt)
}

func (s *AssteroddScene) Render(dt time.Duration) {
	// Animation
	s.World.Systems.AnimationSpriteMatrix.Run()
	s.World.Systems.AnimationPlayer.Run()

	s.World.Systems.SpriteMatrix.Run()
	s.World.Systems.Sprite.Run()
	s.World.Systems.Debug.Run()
	s.World.Systems.AssetLib.Run()
	s.World.Systems.YSort.Run()

	shouldContinue := s.World.Systems.RenderAssterodd.Run(dt)
	if !shouldContinue {
		s.Game.SetShouldDestroy(true)
		return
	}
}

func (s *AssteroddScene) Destroy() {
	s.World.Destroy()
	s.World.Systems.ColliderSystem.Destroy()

	s.World.Systems.SpaceSpawner.Destroy()
	s.World.Systems.SpaceshipIntents.Destroy()

	s.World.Systems.AssteroddSystem.Destroy()

	s.World.Systems.CollisionDetectionBVH.Destroy()
	s.World.Systems.CollisionResolution.Destroy()

	// Animation
	s.World.Systems.AnimationSpriteMatrix.Destroy()
	s.World.Systems.AnimationPlayer.Destroy()

	s.World.Systems.Sprite.Destroy()
	s.World.Systems.SpriteMatrix.Destroy()
	s.World.Systems.YSort.Destroy()

	// RenderAssterodd
	s.World.Systems.Debug.Destroy()
	s.World.Systems.AssetLib.Destroy()
	s.World.Systems.RenderAssterodd.Destroy()
	s.World.Systems.Audio.Destroy()
}

func (s *AssteroddScene) OnEnter() {

}

func (s *AssteroddScene) OnExit() {

}

var _ gomp.AnyScene = (*AssteroddScene)(nil)
