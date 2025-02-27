/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package ecs

import "math"

// GridKey is {cellX, cellY} key for SpatialGrid
type GridKey = [2]int

type GridCell struct {
	entities []Entity
	lookup   map[Entity]int
}

type SpatialGrid struct {
	cells          []GridCell
	keys           []GridKey
	lookupByKey    map[GridKey]int
	lookupByEntity map[Entity]int

	cellSize float64

	updated map[GridKey]struct{}
}

func NewSpatialGrid(cellSize float64) SpatialGrid {
	return SpatialGrid{
		cells:          make([]GridCell, 0, PREALLOC_DEFAULT),
		keys:           make([]GridKey, 0, PREALLOC_DEFAULT),
		lookupByKey:    make(map[GridKey]int, PREALLOC_DEFAULT),
		lookupByEntity: make(map[Entity]int, PREALLOC_DEFAULT),
		cellSize:       cellSize,
	}
}

// GetCell returns the grid cell for a given position.
func (g *SpatialGrid) GetCell(positionX, positionY float64) (*GridCell, GridKey) {
	cellX := int(math.Floor(positionX / g.cellSize))
	cellY := int(math.Floor(positionY / g.cellSize))

	key := GridKey{cellX, cellY}
	cell := &g.cells[g.lookupByKey[key]]
	return cell, key
}

// AddEntity adds an entity to the grid.
func (g *SpatialGrid) AddEntity(entity Entity, positionX, positionY float64) {
	cellX := int(math.Floor(positionX / g.cellSize))
	cellY := int(math.Floor(positionY / g.cellSize))

	key := GridKey{cellX, cellY}

	cellIndex, ok := g.lookupByKey[key]
	if !ok {
		cellIndex = len(g.cells)
		g.cells = append(g.cells, GridCell{
			entities: make([]Entity, 0, PREALLOC_DEFAULT),
			lookup:   make(map[Entity]int, PREALLOC_DEFAULT),
		})
	}
	cell := &g.cells[cellIndex]

	cell.entities = append(cell.entities, entity)
	cell.lookup[entity] = len(cell.entities) - 1

	g.lookupByEntity[entity] = cellIndex
	g.lookupByKey[key] = cellIndex
}

// GetEntities returns the entities in a given cell.
func (g *SpatialGrid) GetEntities(cellX, cellY int) []Entity {
	key := GridKey{cellX, cellY}
	cellIndex, ok := g.lookupByKey[key]
	if !ok {
		return nil
	}
	return g.cells[cellIndex].entities
}

// UpdateEntity updates state of an entity in the grid.
func (g *SpatialGrid) UpdateEntity(entity Entity) {
	g.updated[g.keys[g.lookupByEntity[entity]]] = struct{}{}
}

// RemoveEntity removes an entity from the grid.
func (g *SpatialGrid) RemoveEntity(entity Entity) {
	cellIndex := g.lookupByEntity[entity]
	cell := &g.cells[cellIndex]

	last := len(cell.entities) - 1
	entityIndex := cell.lookup[entity]

	if entityIndex != last {
		cell.entities[entityIndex] = cell.entities[last]
		cell.lookup[cell.entities[last]] = entityIndex
	}

	cell.entities = cell.entities[:last]
	delete(cell.lookup, entity)
	delete(g.lookupByEntity, entity)

	if len(cell.entities) == 0 {
		delete(g.lookupByKey, GridKey{cellIndex, cellIndex})
	}
}

// MoveEntity updates the position of an entity in the grid.
func (g *SpatialGrid) MoveEntity(entity Entity, positionX, positionY float64) {
	oldCellIndex := g.lookupByEntity[entity]
	oldCell := &g.cells[oldCellIndex]

	newCellX := int(math.Floor(positionX / g.cellSize))
	newCellY := int(math.Floor(positionY / g.cellSize))
	newKey := GridKey{newCellX, newCellY}

	// If the entity hasn't changed cells, no need to move it
	if g.keys[oldCellIndex] == newKey {
		g.updated[newKey] = struct{}{}
		return
	}

	// Remove from old cell
	last := len(oldCell.entities) - 1
	entityIndex := oldCell.lookup[entity]
	if entityIndex != last {
		oldCell.entities[entityIndex] = oldCell.entities[last]
		oldCell.lookup[oldCell.entities[last]] = entityIndex
	}
	oldCell.entities = oldCell.entities[:last]
	delete(oldCell.lookup, entity)

	// Add to new cell
	newCellIndex, ok := g.lookupByKey[newKey]
	if !ok {
		newCellIndex = len(g.cells)
		g.cells = append(g.cells, GridCell{
			entities: make([]Entity, 0, PREALLOC_DEFAULT),
			lookup:   make(map[Entity]int, PREALLOC_DEFAULT),
		})
		g.lookupByKey[newKey] = newCellIndex
		g.keys = append(g.keys, newKey)
	}
	newCell := &g.cells[newCellIndex]
	newCell.entities = append(newCell.entities, entity)
	newCell.lookup[entity] = len(newCell.entities) - 1

	// Update entity lookup
	g.lookupByEntity[entity] = newCellIndex
	g.updated[newKey] = struct{}{}
}
