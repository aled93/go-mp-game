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

package gomp

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"time"
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct{}

func (s *RenderSystem) Init() {
	rl.InitWindow(0, 0, "raylib [core] ebiten-ecs - basic window")
}
func (s *RenderSystem) Run(dt time.Duration) bool {
	if rl.WindowShouldClose() {
		return false
	}
	return true
}

func (s *RenderSystem) Destroy() {
	rl.CloseWindow()
}
