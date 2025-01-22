/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"gomp_game/cmd/raylib-ecs/components"
	"gomp_game/pkgs/gomp/ecs"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type spawnController struct {
	pprofEnabled      bool
	slingshotCharging bool
	slingshotStart    rl.Vector2
}

const (
	minHpPercentage = 20
	minMaxHp        = 50000
	maxMaxHp        = 200000
)

func (s *spawnController) Init(world *ecs.World) {}
func (s *spawnController) Update(world *ecs.World) {
	sprites := components.SpriteService.GetManager(world)
	healths := components.HealthService.GetManager(world)
	positions := components.PositionService.GetManager(world)
	rotations := components.RotationService.GetManager(world)
	scales := components.ScaleService.GetManager(world)
	velocities := components.VelocityService.GetManager(world)
	gravemits := components.GravitationEmitterService.GetManager(world)
	gravrecvs := components.GravitationReceiveService.GetManager(world)

	if rl.IsKeyDown(rl.KeySpace) {
		for range rand.Intn(100) {
			if world.Size() > 100_000_000 {
				break
			}

			newCreature := world.CreateEntity("Creature")

			// Adding position component
			t := components.Position{
				X: float32(rand.Int31n(800)),
				Y: float32(rand.Int31n(600)),
			}
			positions.Create(newCreature, t)

			// Adding rotation component
			rotation := components.Rotation{
				Angle: float32(rand.Int31n(360)),
			}
			rotations.Create(newCreature, rotation)

			// Adding scale component
			scale := components.Scale{
				X: 2,
				Y: 2,
			}
			scales.Create(newCreature, scale)

			velocities.Create(newCreature, components.Velocity{
				X: (rand.Float32()*2.0 - 1.0) * 0.0001,
				Y: (rand.Float32()*2.0 - 1.0) * 0.0001,
			})

			gravemits.Create(newCreature, components.GravitationEmitter{})
			gravrecvs.Create(newCreature, components.GravitationReceiver{})

			// Adding HP component
			maxHp := minMaxHp + rand.Int31n(maxMaxHp-minMaxHp)
			hp := int32(float32(maxHp) * float32(minHpPercentage+rand.Int31n(100-minHpPercentage)) / 100)
			h := components.Health{
				Hp:    hp,
				MaxHp: maxHp,
			}
			healths.Create(newCreature, h)

			// Adding sprite component
			c := components.Sprite{
				Origin: rl.Vector2{X: 0.5, Y: 0.5},
			}
			sprites.Create(newCreature, c)
		}
	}

	mpos := rl.GetMousePosition()

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		star := world.CreateEntity("Star")

		positions.Create(star, components.Position{
			X: mpos.X,
			Y: mpos.Y,
		})

		rotations.Create(star, components.Rotation{})
		scales.Create(star, components.Scale{
			X: 2.0,
			Y: 2.0,
		})

		velocities.Create(star, components.Velocity{
			X: 0.0,
			Y: 0.0,
		})

		gravemits.Create(star, components.GravitationEmitter{})
		gravrecvs.Create(star, components.GravitationReceiver{})

		healths.Create(star, components.Health{
			Hp:    10000,
			MaxHp: 10000,
		})

		sprites.Create(star, components.Sprite{
			Origin: rl.NewVector2(0.5, 0.5),
		})
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		s.slingshotCharging = true
		s.slingshotStart = mpos
	}
	if rl.IsMouseButtonReleased(rl.MouseButtonRight) && s.slingshotCharging {
		s.slingshotCharging = false
		vel := rl.Vector2Scale(rl.Vector2Subtract(mpos, s.slingshotStart), 0.01)
		// frc := rl.Vector2Distance(mpos, s.slingshotStart)

		star := world.CreateEntity("Star")

		positions.Create(star, components.Position{
			X: mpos.X,
			Y: mpos.Y,
		})

		rotations.Create(star, components.Rotation{})
		scales.Create(star, components.Scale{
			X: 2.0,
			Y: 2.0,
		})

		velocities.Create(star, components.Velocity{
			X: vel.X,
			Y: vel.Y,
		})

		gravemits.Create(star, components.GravitationEmitter{})
		gravrecvs.Create(star, components.GravitationReceiver{})

		healths.Create(star, components.Health{
			Hp:    10000,
			MaxHp: 10000,
		})

		sprites.Create(star, components.Sprite{
			Origin: rl.NewVector2(0.5, 0.5),
		})
	}
}
func (s *spawnController) FixedUpdate(world *ecs.World) {
}

func (s *spawnController) Destroy(world *ecs.World) {}
