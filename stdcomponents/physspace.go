package stdcomponents

import (
	"gomp/examples/new-api/ids"
	"gomp/pkg/collision"
	"gomp/pkg/ecs"
)

type PhysSpace struct {
	*collision.Space2D
	InitParams collision.SpaceCreateParams
}

type PhysSpaceComponentManager = ecs.SharedComponentManager[PhysSpace]

func NewPhysSpaceComponentManager() PhysSpaceComponentManager {
	return ecs.NewSharedComponentManager[PhysSpace](ecs.ComponentId(ids.DefaultPhysSpaceSharedComponentId))
}
