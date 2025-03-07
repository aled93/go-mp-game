package stdsystems

import (
	"gomp/pkg/collision"
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

type PhysSpaceSyncSystem struct {
	PhysSpaces *stdcomponents.PhysSpaceComponentManager
}

func (sys *PhysSpaceSyncSystem) Init() {
}

func (sys *PhysSpaceSyncSystem) Run() {
	sys.PhysSpaces.Each(func(ent ecs.Entity, physSpace *stdcomponents.PhysSpace) bool {
		if physSpace.Space2D == nil {
			physSpace.Space2D = collision.NewSpace(&physSpace.InitParams)
		}

		physSpace.DebugDraw(0)

		return true
	})
}

func (sys *PhysSpaceSyncSystem) Destroy() {
}
