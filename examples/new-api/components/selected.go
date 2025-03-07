package components

import "gomp/pkg/ecs"

type Selected struct{}

type SelectedComponentManager = ecs.ComponentManager[Selected]

func NewSelectedComponentManager() SelectedComponentManager {
	return ecs.NewComponentManager[Selected](SelectedComponentId)
}
