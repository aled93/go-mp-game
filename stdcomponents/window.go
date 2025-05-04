package stdcomponents

import (
	"gomp/pkg/ecs"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

type Window struct {
	Handle *sdl.Window
}

type WindowComponentManager = ecs.ComponentManager[Window]

func NewWindowComponentManager() WindowComponentManager {
	return ecs.NewComponentManager[Window](WindowComponentId)
}
