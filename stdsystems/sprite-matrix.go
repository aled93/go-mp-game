/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

func NewSpriteMatrixSystem() SpriteMatrixSystem {
	return SpriteMatrixSystem{}
}

// SpriteMatrixSystem is a system that prepares SpriteSheet to be rendered
type SpriteMatrixSystem struct {
	SpriteMatrixes  *stdcomponents.SpriteMatrixComponentManager
	RLTexturePros   *stdcomponents.RLTextureProComponentManager
	AnimationStates *stdcomponents.AnimationStateComponentManager
}

func (s *SpriteMatrixSystem) Init() {}
func (s *SpriteMatrixSystem) Run() {
	s.SpriteMatrixes.EachEntity(func(entity ecs.Entity) bool {
		spriteMatrix := s.SpriteMatrixes.Get(entity)    //
		animationState := s.AnimationStates.Get(entity) //

		frame := spriteMatrix.Animations[*animationState].Frame

		tr := s.RLTexturePros.Get(entity)
		if tr == nil {
			s.RLTexturePros.Create(entity, stdcomponents.RLTexturePro{
				Texture: spriteMatrix.Texture, //
				Frame:   frame,                //
				Origin:  spriteMatrix.Origin,
				Dest:    spriteMatrix.Dest, //
			})
		} else {
			// Run spriteRender
			tr.Dest = spriteMatrix.Dest
			tr.Frame = frame
		}
		return true
	})
}
func (s *SpriteMatrixSystem) Destroy() {}
