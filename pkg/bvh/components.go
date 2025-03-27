package bvh

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
)

type treeComponent struct {
	Entity ecs.Entity
	AABB   *stdcomponents.AABB
}
