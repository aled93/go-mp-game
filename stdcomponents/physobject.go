package stdcomponents

import (
	"gomp/pkg/collision"
	"gomp/pkg/ecs"
)

type PhysObject struct {
	*collision.Object
}

type PhysObjectComponentManager = ecs.ComponentManager[PhysObject]

func NewPhysObjectComponentManager() PhysObjectComponentManager {
	return ecs.NewComponentManager[PhysObject](PhysObjectComponentId)
}
