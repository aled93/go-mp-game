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
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RLTexturePro struct {
	Texture  *rl.Texture2D
	Frame    util.Rect
	Origin   util.Vec2
	Tint     color.RGBA
	Dest     util.Rect
	Rotation float32
}

func (t *RLTexturePro) Rect() util.Rect {
	return util.NewRectFromOriginSize(t.Dest.Mins.Subtract(t.Origin), t.Dest.Size())
}

type RLTextureProComponentManager = ecs.ComponentManager[RLTexturePro]

func NewRlTextureProComponentManager() RLTextureProComponentManager {
	return ecs.NewComponentManager[RLTexturePro](RLTextureProComponentId)
}
