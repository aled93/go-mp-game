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

package sprites

import (
	"gomp/examples/new-api/assets"
	"gomp/pkg/util"
	"gomp/stdcomponents"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var PlayerSpriteMatrix = stdcomponents.SpriteMatrix{
	Texture: assets.Textures.Get("milansheet.png"),
	Origin:  util.NewVec2(0.5, 0.5),
	Dest:    util.NewRectFromOriginSize(util.NewVec2(0, 0), util.NewVec2(96, 128)),
	FPS:     12,
	Animations: []stdcomponents.SpriteMatrixAnimation{
		{
			Name:        "idle",
			Frame:       rl.Rectangle{X: 0, Y: 0, Width: 96, Height: 128},
			NumOfFrames: 1,
			Vertical:    false,
			Loop:        true,
		},
		{
			Name:        "walk",
			Frame:       rl.Rectangle{X: 0, Y: 512, Width: 96, Height: 128},
			NumOfFrames: 8,
			Vertical:    false,
			Loop:        true,
		},
		{
			Name:        "jump",
			Frame:       rl.Rectangle{X: 96, Y: 0, Width: 96, Height: 128},
			NumOfFrames: 1,
			Vertical:    false,
			Loop:        false,
		},
	},
}
