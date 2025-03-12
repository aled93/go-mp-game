package components

import (
	"gomp/pkg/ecs"
)

type PickupPower string

const (
	PickupPower_Hp PickupPower = "hp"
	// PickupPower_ PickupPower = "hp"
)

type Pickup struct {
	Power  PickupPower
	Amount int
}

type PickupComponentManager = ecs.ComponentManager[Pickup]

func NewPickupComponentManager() PickupComponentManager {
	return ecs.NewComponentManager[Pickup](PickupComponentId)
}
