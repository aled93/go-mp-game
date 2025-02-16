/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp/examples/raylib-ecs/components"
	ecs2 "gomp/pkg/ecs"
	"time"
)

type animationSpriteMatrixController struct{}

func (s *animationSpriteMatrixController) Init(world *ecs2.World) {}
func (s *animationSpriteMatrixController) Update(world *ecs2.World) {
	animationPlayers := components.AnimationPlayerService.GetManager(world)
	animationStates := components.AnimationStateService.GetManager(world)
	spriteMatrixes := components.SpriteMatrixService.GetManager(world)

	animationPlayers.AllParallel(func(e ecs2.Entity, animationPlayer *components.AnimationPlayer) bool {
		spriteMatrix := spriteMatrixes.Get(e)
		if spriteMatrix == nil {
			return true
		}

		animationStatePtr := animationStates.Get(e)
		if animationStatePtr == nil {
			return true
		}
		animationState := *animationStatePtr

		if animationPlayer.State == animationState && animationPlayer.IsInitialized == true {
			return true
		}

		currentAnimation := spriteMatrix.Animations[animationState]

		animationPlayer.First = 0
		animationPlayer.Current = 0
		animationPlayer.Last = currentAnimation.NumOfFrames - 1
		animationPlayer.Loop = currentAnimation.Loop
		animationPlayer.Vertical = currentAnimation.Vertical
		animationPlayer.FrameDuration = time.Second / time.Duration(spriteMatrix.FPS)
		animationPlayer.State = animationState
		animationPlayer.Speed = 1
		animationPlayer.IsInitialized = true

		return true
	})
}
func (s *animationSpriteMatrixController) FixedUpdate(world *ecs2.World) {}
func (s *animationSpriteMatrixController) Destroy(world *ecs2.World)     {}
