package systems

import (
	"gomp/examples/raylib-ecs/components"
	"gomp/pkg/ecs"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type localInputController struct {
	selector ecs.Selector[struct {
		*components.LocalInput
		Input *components.InputIntent
	}]
}

func (s *localInputController) Init(world *ecs.World) {
	world.RegisterSelector(&s.selector)
}

func (s *localInputController) Update(world *ecs.World) {
	for e := range s.selector.All() {
		if rl.IsKeyDown(rl.KeyA) {
			e.Input.Move.X = -1.0
		} else if rl.IsKeyDown(rl.KeyD) {
			e.Input.Move.X = +1.0
		} else {
			e.Input.Move.X = 0.0
		}
		if rl.IsKeyDown(rl.KeyW) {
			e.Input.Move.Y = -1.0
		} else if rl.IsKeyDown(rl.KeyS) {
			e.Input.Move.Y = +1.0
		} else {
			e.Input.Move.Y = 0.0
		}
	}
}

func (s *localInputController) FixedUpdate(world *ecs.World) {}
func (s *localInputController) Destroy(world *ecs.World)     {}
