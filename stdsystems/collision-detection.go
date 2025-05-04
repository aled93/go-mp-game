/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

<- Фрея Donated 2 000 RUB

Thank you for your support!
*/

package stdsystems

import (
	"github.com/negrel/assert"
	gjk "gomp/pkg/collision"
	"gomp/pkg/core"
	"gomp/pkg/ecs"
	"gomp/pkg/worker"
	"gomp/stdcomponents"
	"gomp/vectors"
	"time"
)

func NewCollisionDetectionSystem() CollisionDetectionSystem {
	return CollisionDetectionSystem{}
}

type CollisionDetectionSystem struct {
	EntityManager                      *ecs.EntityManager
	Positions                          *stdcomponents.PositionComponentManager
	Rotations                          *stdcomponents.RotationComponentManager
	Scales                             *stdcomponents.ScaleComponentManager
	GenericCollider                    *stdcomponents.GenericColliderComponentManager
	BoxColliders                       *stdcomponents.BoxColliderComponentManager
	CircleColliders                    *stdcomponents.CircleColliderComponentManager
	PolygonColliders                   *stdcomponents.PolygonColliderComponentManager
	Collisions                         *stdcomponents.CollisionComponentManager
	SpatialIndex                       *stdcomponents.SpatialHashComponentManager
	AABB                               *stdcomponents.AABBComponentManager
	ColliderSleepStateComponentManager *stdcomponents.ColliderSleepStateComponentManager
	BvhTreeComponentManager            *stdcomponents.BvhTreeComponentManager
	CollisionGridComponentManager      *stdcomponents.CollisionGridComponentManager
	CollisionChunkComponentManager     *stdcomponents.CollisionChunkComponentManager
	CollisionCellComponentManager      *stdcomponents.CollisionCellComponentManager
	Engine                             *core.Engine

	collisionEventAcc []ecs.PagedArray[CollisionEvent]
	activeCollisions  map[CollisionPair]ecs.Entity // Maps collision pairs to proxy entities
	currentCollisions map[CollisionPair]struct{}
	gridLookup        map[stdcomponents.CollisionLayer]*stdcomponents.CollisionGrid
}

func (s *CollisionDetectionSystem) Init() {
	s.gridLookup = make(map[stdcomponents.CollisionLayer]*stdcomponents.CollisionGrid)
	s.collisionEventAcc = make([]ecs.PagedArray[CollisionEvent], s.Engine.Pool().NumWorkers())
	for i := 0; i < len(s.collisionEventAcc); i++ {
		s.collisionEventAcc[i] = ecs.NewPagedArray[CollisionEvent]()
	}
	s.activeCollisions = make(map[CollisionPair]ecs.Entity)
}

func (s *CollisionDetectionSystem) Run(dt time.Duration) {
	s.CollisionGridComponentManager.EachComponent()(func(grid *stdcomponents.CollisionGrid) bool {
		s.gridLookup[grid.Layer] = grid
		return true
	})

	s.currentCollisions = make(map[CollisionPair]struct{})

	if s.GenericCollider.Len() == 0 {
		return
	}

	s.GenericCollider.ProcessEntities(func(entity ecs.Entity, workerId worker.WorkerId) {
		potentialEntities := s.broadPhase(entity, make([]ecs.Entity, 0, 64))
		if len(potentialEntities) == 0 {
			return
		}

		s.narrowPhase(entity, potentialEntities, workerId)
	})

	s.registerCollisionEvents()
	s.processExitStates()

	for i := 0; i < len(s.collisionEventAcc); i++ {
		s.collisionEventAcc[i].Reset()
	}
}

func (s *CollisionDetectionSystem) Destroy() {
	s.collisionEventAcc = nil
	s.activeCollisions = nil
	s.currentCollisions = nil
	s.gridLookup = nil
}

func (s *CollisionDetectionSystem) broadPhase(entityA ecs.Entity, result []ecs.Entity) []ecs.Entity {
	colliderA := s.GenericCollider.GetUnsafe(entityA)
	if colliderA.AllowSleep {
		if s.ColliderSleepStateComponentManager.Has(entityA) {
			return result
		}
	}

	aabbPtr := s.AABB.GetUnsafe(entityA)
	assert.NotNil(aabbPtr)
	aabb := *aabbPtr

	cells := make([]ecs.Entity, 0, 64)
	// Iterate through all trees
	for index := range s.gridLookup {
		grid := s.gridLookup[index]
		layer := grid.Layer

		// Check if mask includes this layer
		if !colliderA.Mask.HasLayer(layer) {
			continue
		}

		// Traverse this BVH tree for potential collisions
		cells = grid.Query(aabb, cells)
	}

	for _, cellEntityId := range cells {
		cell := s.CollisionCellComponentManager.GetUnsafe(cellEntityId)
		assert.NotNil(cell)

		result = append(result, cell.Members.Members...)
	}

	return result
}

func (s *CollisionDetectionSystem) narrowPhase(entityA ecs.Entity, potentialEntities []ecs.Entity, workerId worker.WorkerId) {
	for _, entityB := range potentialEntities {
		if entityA == entityB {
			continue
		}

		colliderA := s.GenericCollider.GetUnsafe(entityA)
		colliderB := s.GenericCollider.GetUnsafe(entityB)
		posA := s.Positions.GetUnsafe(entityA)
		posB := s.Positions.GetUnsafe(entityB)
		scaleA := s.Scales.GetUnsafe(entityA)
		scaleB := s.Scales.GetUnsafe(entityB)
		rotA := s.Rotations.GetUnsafe(entityA)
		rotB := s.Rotations.GetUnsafe(entityB)
		transformA := stdcomponents.Transform2d{
			Position: posA.XY,
			Rotation: rotA.Angle,
			Scale:    scaleA.XY,
		}
		transformB := stdcomponents.Transform2d{
			Position: posB.XY,
			Rotation: rotB.Angle,
			Scale:    scaleB.XY,
		}

		circleA := s.CircleColliders.GetUnsafe(entityA)
		circleB := s.CircleColliders.GetUnsafe(entityB)
		if circleA != nil && circleB != nil {
			radiusA := circleA.Radius * scaleA.XY.X
			radiusB := circleB.Radius * scaleB.XY.X
			if transformA.Position.Distance(transformB.Position) < radiusA+radiusB {
				s.collisionEventAcc[workerId].Append(CollisionEvent{
					entityA:  entityA,
					entityB:  entityB,
					position: transformA.Position,
					normal:   transformB.Position.Sub(transformA.Position).Normalize(),
					depth:    radiusA + radiusB - transformB.Position.Distance(transformA.Position),
				})
			}
			continue
		}

		// GJK strategy
		colA := s.getGjkCollider(colliderA, entityA)
		colB := s.getGjkCollider(colliderB, entityB)
		// First detect collision using GJK
		test := gjk.New()
		if !test.CheckCollision(colA, colB, transformA, transformB) {
			continue
		}

		// If collision detected, get penetration details using EPA
		normal, depth := test.EPA(colA, colB, transformA, transformB)
		position := posA.XY.Add(posB.XY.Sub(posA.XY))
		s.collisionEventAcc[workerId].Append(CollisionEvent{
			entityA:  entityA,
			entityB:  entityB,
			position: position,
			normal:   normal,
			depth:    depth,
		})
	}
}
func (s *CollisionDetectionSystem) registerCollisionEvents() {
	for i := range s.collisionEventAcc {
		events := &s.collisionEventAcc[i]
		events.EachData()(func(event *CollisionEvent) bool {
			pair := CollisionPair{event.entityA, event.entityB}
			s.currentCollisions[pair] = struct{}{}
			displacement := event.normal.Scale(event.depth)
			pos := event.position.Add(displacement)

			if _, exists := s.activeCollisions[pair]; !exists {
				proxy := s.EntityManager.Create()
				s.Collisions.Create(proxy, stdcomponents.Collision{
					E1:     pair.E1,
					E2:     pair.E2,
					State:  stdcomponents.CollisionStateEnter,
					Normal: event.normal,
					Depth:  event.depth,
				})

				s.Positions.Create(proxy, stdcomponents.Position{
					XY: vectors.Vec2{
						X: pos.X, Y: pos.Y,
					}})
				s.activeCollisions[pair] = proxy
			} else {
				proxy := s.activeCollisions[pair]
				collision := s.Collisions.GetUnsafe(proxy)
				position := s.Positions.GetUnsafe(proxy)
				collision.State = stdcomponents.CollisionStateStay
				collision.Depth = event.depth
				collision.Normal = event.normal
				position.XY.X = pos.X
				position.XY.Y = pos.Y
			}
			return true
		})
		events.Reset()
	}
}

func (s *CollisionDetectionSystem) processExitStates() {
	for pair, proxy := range s.activeCollisions {
		if _, exists := s.currentCollisions[pair]; !exists {
			collision := s.Collisions.GetUnsafe(proxy)
			if collision.State == stdcomponents.CollisionStateExit {
				delete(s.activeCollisions, pair)
				s.EntityManager.Delete(proxy)
			} else {
				collision.State = stdcomponents.CollisionStateExit
			}
		}
	}
}
func (s *CollisionDetectionSystem) getGjkCollider(collider *stdcomponents.GenericCollider, entity ecs.Entity) gjk.AnyCollider {
	switch collider.Shape {
	case stdcomponents.BoxColliderShape:
		return s.BoxColliders.GetUnsafe(entity)
	case stdcomponents.CircleColliderShape:
		return s.CircleColliders.GetUnsafe(entity)
	case stdcomponents.PolygonColliderShape:
		return s.PolygonColliders.GetUnsafe(entity)
	default:
		panic("unsupported collider shape")
	}
	return nil
}
