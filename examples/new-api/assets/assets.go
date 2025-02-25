/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package assets

import (
	"embed"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
	"gomp"
	"image/png"
	"log"
)

//go:embed *.png
var fs embed.FS

var Textures = gomp.CreateAssetLibrary(
	func(path string) rl.Texture2D {
		assert.True(rl.IsWindowReady(), "Window is not initialized")

		fmt.Print()
		file, err := fs.Open(path)
		if err != nil {
			log.Panic("Error opening file")
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			log.Panic("Error decoding file")
		}

		rlImg := rl.NewImageFromImage(img)
		return rl.LoadTextureFromImage(rlImg)
	},
	func(path string, asset *rl.Texture2D) {
		assert.True(rl.IsWindowReady(), "Window is not initialized")
		rl.UnloadTexture(*asset)
	},
)
