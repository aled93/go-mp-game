package systems

import (
	"gomp/examples/raylib-ecs/components"
	"gomp/pkg/ecs"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const dist_reached_target = 10.0
const chill_duration = 100

type botRoamingController struct {
	selector ecs.Selector[struct {
		*components.BotRoamer
		Position *components.Position
		Input    *components.InputIntent
	}]
}

func (s *botRoamingController) Init(world *ecs.World) {
	world.RegisterSelector(&s.selector)
}

func (s *botRoamingController) Update(world *ecs.World) {
	for e := range s.selector.All() {
		botPos := rl.NewVector2(e.Position.X, e.Position.Y)

		e.Input.Move = rl.Vector2{}

		if e.BotRoamer.Chilling {
			e.BotRoamer.ChillDuration -= 1

			if e.BotRoamer.ChillDuration <= 0 {
				println("enough chill, need to go")
				e.BotRoamer.Chilling = false
				e.BotRoamer.RoamTarget.X = botPos.X + ((rand.Float32()*2.0 - 1.0) * 200.0)
				e.BotRoamer.RoamTarget.Y = botPos.Y + ((rand.Float32()*2.0 - 1.0) * 200.0)
			}
		} else {
			if rl.Vector2DistanceSqr(botPos, e.RoamTarget) < dist_reached_target*dist_reached_target {
				println("reached target, need to chill")
				e.BotRoamer.Chilling = true
				e.BotRoamer.ChillDuration = chill_duration
			} else {
				e.Input.Move = rl.Vector2Normalize(rl.Vector2Subtract(e.RoamTarget, botPos))
				rl.DrawLine(int32(botPos.X), int32(botPos.Y), int32(e.RoamTarget.X), int32(e.RoamTarget.Y), rl.Lime)
			}
		}
	}
}

func (s *botRoamingController) FixedUpdate(world *ecs.World) {}
func (s *botRoamingController) Destroy(world *ecs.World)     {}
