/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

const (
	batchSize = 1 << 13 // Maximum batch size supported by Raylib
)

func NewRenderSystem() RenderSystem {
	return RenderSystem{}
}

type RenderSystem struct {
	EntityManager  *ecs.EntityManager
	TextureRenders *stdcomponents.TextureRenderComponentManager
	Positions      *stdcomponents.PositionComponentManager
	camera         rl.Camera2D
	trBatch        []stdcomponents.TextureRender
}

func (s *RenderSystem) Init() {
	rl.InitWindow(1024, 768, "raylib [core] ebiten-ecs - basic window")
	//InitWindow(1024, 768, "raylib [core] ebiten-ecs - basic window")

	s.camera = rl.Camera2D{
		Target:   rl.NewVector2(0, 0),
		Offset:   rl.NewVector2(0, 0),
		Rotation: 0,
		Zoom:     1,
	}
}

func (s *RenderSystem) Run() bool {
	if rl.WindowShouldClose() {
		return false
	}

	textureLen := s.TextureRenders.Len()
	if len(s.trBatch) < textureLen {
		s.trBatch = make([]stdcomponents.TextureRender, textureLen)
	}
	s.TextureRenders.RawComponents(s.trBatch)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	for batch := 0; batch < textureLen; batch += batchSize {
		endBatch := batch + batchSize
		if endBatch > textureLen {
			endBatch = textureLen
		}

		rl.BeginMode2D(s.camera)
		for i := batch; i < endBatch; i++ {
			tr := &s.trBatch[i]
			txt := *tr.Texture
			rl.DrawTexturePro(txt, tr.Frame, tr.Dest, tr.Origin, tr.Rotation, tr.Tint)
		}
		rl.EndMode2D()
	}

	rl.DrawRectangle(0, 0, 200, 60, rl.DarkBrown)
	rl.DrawFPS(10, 10)
	rl.DrawText(fmt.Sprintf("%d entities", s.EntityManager.Size()), 10, 30, 20, rl.RayWhite)

	rl.EndDrawing()

	return true
}

func (s *RenderSystem) Destroy() {
	rl.CloseWindow()
}
