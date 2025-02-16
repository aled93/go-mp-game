/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package ecs

import (
	"image/color"
	"math/rand"
	"testing"
)

type pixel struct {
	x      int32
	y      int32
	hp     int32
	color  color.RGBA
	breath bool
}

var pixelComponentType = CreateComponentService[pixel](1)

// Commonly used functions in both benchmarks.
func PrepareWorld(description string, system AnySystemServicePtr) *World {
	world := CreateWorld(description)

	world.RegisterComponentServices(
		&pixelComponentType,
	)
	world.RegisterSystems().
		Parallel(
			system,
		)

	return world
}

func InitPixelComponent(pixelComponent *ComponentManager[pixel], world *World) {
	pixelComponent = pixelComponentType.GetManager(world)
	determRand := rand.New(rand.NewSource(42))

	for i := range 1000 {
		for j := range 1000 {
			newPixel := world.CreateEntity("Pixel")

			randomGreen := uint8(135 / (determRand.Intn(10) + 1))
			randomBlue := uint8(135 / (determRand.Intn(10) + 1))

			randomColor := color.RGBA{
				G: randomGreen,
				B: randomBlue,
				A: 255,
			}
			pixelComponent.Create(newPixel, pixel{
				x:     int32(j),
				y:     int32(i),
				hp:    100,
				color: randomColor,
			})
		}
	}
}

type pixelSystem struct {
	pixelComponent ComponentManager[pixel]
}

func (s *pixelSystem) Init(world *World) {
	InitPixelComponent(&s.pixelComponent, world)
}

func (s *pixelSystem) Destroy(world *World) {}

func (s *pixelSystem) Update(world *World) {}
func (s *pixelSystem) FixedUpdate(world *World) {
	for pixel := range s.pixelComponent.AllData {
		// Note: was not extracted to separate function to simulate
		// real-world interaction between range loop and inner code.
		color := &pixel.color

		if pixel.breath {
			if color.G < 135 {
				color.G++
			} else {
				pixel.hp++
			}
			if color.B < 135 {
				color.B++
			} else {
				pixel.hp++
			}
		} else {
			if color.G > 0 {
				color.G--
			} else {
				pixel.hp--
			}
			if color.B > 0 {
				color.B--
			} else {
				pixel.hp--
			}
		}

		if pixel.hp <= 0 {
			pixel.breath = true
		} else if pixel.hp >= 100 {
			pixel.breath = false
		}
	}
}

// Direct call iteration type
type pixelSystemDirectCall struct {
	pixelComponent ComponentManager[pixel]
}

func (s *pixelSystemDirectCall) Init(world *World) {
	InitPixelComponent(&s.pixelComponent, world)
}

func (s *pixelSystemDirectCall) Destroy(world *World) {}
func (s *pixelSystemDirectCall) Update(world *World)  {}
func (s *pixelSystemDirectCall) FixedUpdate(world *World) {
	s.pixelComponent.AllDataParallel(func(pixel *pixel) bool {
		color := &pixel.color

		if pixel.breath {
			if color.G < 135 {
				color.G++
			} else {
				pixel.hp++
			}
			if color.B < 135 {
				color.B++
			} else {
				pixel.hp++
			}
		} else {
			if color.G > 0 {
				color.G--
			} else {
				pixel.hp--
			}
			if color.B > 0 {
				color.B--
			} else {
				pixel.hp--
			}
		}

		if pixel.hp <= 0 {
			pixel.breath = true
		} else if pixel.hp >= 100 {
			pixel.breath = false
		}
		return true
	})
}

// Note: amount of memory allocated changes between tests even with deterministic rand.
// Observed range 918063 B/op - 1108007 B/op
func BenchmarkRangeIteration(b *testing.B) {
	pixelSys := CreateSystemService(new(pixelSystem))
	world := PrepareWorld("range iteration", &pixelSys)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		world.runSystemFunction(SystemFunctionFixedUpdate)
	}
}

// Note: amount of memory allocated changes between tests even with deterministic rand.
// Observed range 868437 B/op - 1047789 B/op
func BenchmarkDirectCallIteration(b *testing.B) {
	pixelSys := CreateSystemService(new(pixelSystemDirectCall))
	world := PrepareWorld("direct call iteration", &pixelSys)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		world.runSystemFunction(SystemFunctionFixedUpdate)
	}
}
