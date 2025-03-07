package stdcomponents

import "gomp/pkg/ecs"

type Collider struct {
	Size float64
}

type ColliderComponentManager = ecs.ComponentManager[Collider]

func NewColliderComponentManager() ColliderComponentManager {
	return ecs.NewComponentManager[Collider](ColliderComponentId)
}
