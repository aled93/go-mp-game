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

package stdcomponents

import (
	"gomp/pkg/ecs"
	"gomp/pkg/util"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SpriteMatrixAnimation struct {
	Name        string
	Frame       rl.Rectangle
	NumOfFrames uint8
	Vertical    bool
	Loop        bool
}

type SpriteMatrix struct {
	Texture    *rl.Texture2D
	Origin     util.Vec2
	Dest       util.Rect
	FPS        int32
	Animations []SpriteMatrixAnimation
	Rotation   util.Radians
}

type SpriteMatrixComponentManager = ecs.SharedComponentManager[SpriteMatrix]

func NewSpriteMatrixComponentManager() SpriteMatrixComponentManager {
	return ecs.NewSharedComponentManager[SpriteMatrix](SpriteMatrixComponentId)
}
