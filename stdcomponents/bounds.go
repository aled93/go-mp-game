package stdcomponents

import "gomp/pkg/ecs"

type Bounds struct {
	Width, Height float32
}

type BoundsComponentManager = ecs.ComponentManager[Bounds]

func NewBoundsComponentManager() BoundsComponentManager {
	return ecs.NewComponentManager[Bounds](BoundsComponentId)
}
