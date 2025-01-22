/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package spatial

import (
	"slices"
)

type QuadTree2D[N any, E any] struct {
	root              *QuadNode[N, E]
	maxDepth          int
	maxCellPopulation int
	freeId            uint64
}

func NewQuadTree2D[N any, E any](maxX, maxY float64, maxDepth, maxCellPopulation int) *QuadTree2D[N, E] {
	if maxX <= 0.0 || maxY <= 0.0 {
		panic("maxX and maxY must be positive")
	}
	if maxDepth < 1 || maxDepth >= 64 {
		panic("maxDepth must be in range (1..63)")
	}
	if maxCellPopulation < 1 {
		panic("maxCellPopulation must be positive")
	}

	return &QuadTree2D[N, E]{
		root: &QuadNode[N, E]{
			maxX: maxX,
			maxY: maxY,
		},
		maxDepth:          maxDepth,
		maxCellPopulation: maxCellPopulation,
	}
}

func (t *QuadTree2D[N, E]) Root() *QuadNode[N, E] {
	return t.root
}

func (t *QuadTree2D[N, E]) AddPoint(x, y float64) (ent *Entity[N, E], ok bool) {
	ent = &Entity[N, E]{
		x: x,
		y: y,
	}
	return ent, t.addPoint(ent)
}

func (t *QuadTree2D[N, E]) addPoint(ent *Entity[N, E]) bool {
	if ent.x < t.root.minX || ent.y < t.root.minY || ent.x >= t.root.maxX || ent.y >= t.root.maxY {
		return false
	}

	targetNode := t.root.findDeepestNodeAt(ent.x, ent.y)
	targetNode.ents = append(targetNode.ents, ent)
	ent.node = targetNode
	// targetNode.splitRecursively(t.maxCellPopulation, t.maxDepth)

	return true
}

func (t *QuadTree2D[N, E]) Remove(ent *Entity[N, E]) bool {
	idx := slices.IndexFunc(ent.node.ents, func(e *Entity[N, E]) bool {
		return e == ent
	})
	if idx < 0 {
		return false
	}

	ent.node.ents[idx] = nil
	// entCount := len(ent.node.ents)
	// ent.node.ents[idx] = ent.node.ents[entCount-1]
	// ent.node.ents = ent.node.ents[:entCount-1]
	// ent.node.mergeRecursively(t.maxCellPopulation)

	return true
}

func (t *QuadTree2D[N, E]) UpdatePosition(ent *Entity[N, E], newX, newY float64) bool {
	if newX >= ent.node.minX && newY >= ent.node.minY && newX < ent.node.maxX && newY < ent.node.maxY {
		ent.x = newX
		ent.y = newY

		return true
	}

	t.Remove(ent)

	ent.x = newX
	ent.y = newY
	return t.addPoint(ent)
}

func (t *QuadTree2D[N, E]) GetCellAt(x, y float64) (xmin, ymin, xmax, ymax float64, entCount, depth int, ok bool) {
	if x < t.root.minX || y < t.root.minY || x >= t.root.maxX || y >= t.root.maxY {
		return xmin, ymin, xmax, ymax, entCount, depth, false
	}

	n := t.root.findDeepestNodeAt(x, y)
	return n.minX, n.minY, n.maxX, n.maxY, len(n.ents), n.depth, true
}

func (t *QuadTree2D[N, E]) Maintain() {
	t.root.removeMarkedEntitiesRecursively()
	t.root.splitRecursively(t.maxCellPopulation, t.maxDepth)
	t.root.mergeRecursively(t.maxCellPopulation)
}

type QuadNode[N any, E any] struct {
	parent     *QuadNode[N, E]
	childs     [4]*QuadNode[N, E]
	minX, minY float64
	maxX, maxY float64
	midX, midY float64
	depth      int
	ents       []*Entity[N, E]
	userData   N
}

func (n *QuadNode[N, E]) Parent() *QuadNode[N, E] {
	return n.parent
}

func (n *QuadNode[N, E]) Childs() [4]*QuadNode[N, E] {
	return n.childs
}

func (n *QuadNode[N, E]) Entities() []*Entity[N, E] {
	return n.ents
}

func (n QuadNode[N, E]) Bounds() (minX, minY, maxX, maxY float64) {
	return n.minX, n.minY, n.maxX, n.maxY
}

func (n *QuadNode[N, E]) UserData() *N {
	return &n.userData
}

func (n *QuadNode[N, E]) isLeaf() bool {
	return n.ents != nil
}

func (n *QuadNode[N, E]) numSubleafs() (num int) {
	for _, child := range n.childs {
		if child != nil && child.isLeaf() {
			num++
		}
	}

	return num
}

func (n *QuadNode[N, E]) splitRecursively(maxCellEntityCount, maxDepth int) {
	for _, child := range n.childs {
		if child != nil {
			child.splitRecursively(maxCellEntityCount, maxDepth)
		}
	}

	if n.depth >= maxDepth || n.ents == nil || len(n.ents) <= maxCellEntityCount {
		return
	}

	n.split()
}

func (n *QuadNode[N, E]) split() {
	n.childs[0] = &QuadNode[N, E]{
		parent: n,
		minX:   n.minX,
		minY:   n.minY,
		maxX:   n.midX,
		maxY:   n.midY,
		depth:  n.depth + 1,
	}
	n.childs[0].calcMid()
	n.childs[1] = &QuadNode[N, E]{
		parent: n,
		minX:   n.midX,
		minY:   n.minY,
		maxX:   n.maxX,
		maxY:   n.midY,
		depth:  n.depth + 1,
	}
	n.childs[1].calcMid()
	n.childs[2] = &QuadNode[N, E]{
		parent: n,
		minX:   n.minX,
		minY:   n.midY,
		maxX:   n.midX,
		maxY:   n.maxY,
		depth:  n.depth + 1,
	}
	n.childs[2].calcMid()
	n.childs[3] = &QuadNode[N, E]{
		parent: n,
		minX:   n.midX,
		minY:   n.midY,
		maxX:   n.maxX,
		maxY:   n.maxY,
		depth:  n.depth + 1,
	}
	n.childs[3].calcMid()

	for _, ent := range n.ents {
		if ent.y < n.midY {
			if ent.x < n.midX {
				ent.node = n.childs[0]
				n.childs[0].ents = append(n.childs[0].ents, ent)
			} else {
				ent.node = n.childs[1]
				n.childs[1].ents = append(n.childs[1].ents, ent)
			}
		} else {
			if ent.x < n.midX {
				ent.node = n.childs[2]
				n.childs[2].ents = append(n.childs[2].ents, ent)
			} else {
				ent.node = n.childs[3]
				n.childs[3].ents = append(n.childs[3].ents, ent)
			}
		}
	}

	clear(n.ents)
	n.ents = nil
}

func (n *QuadNode[N, E]) mergeRecursively(maxCellEntityCount int) {
	// if !n.isLeaf() {
	// 	for _, child := range n.childs {
	// 		if child != nil {
	// 			child.mergeRecursively(maxCellEntity[N, E]Count)
	// 		}
	// 	}
	// }

	// if len(n.childs[0].ents)+len(n.childs[1].ents)+len(n.childs[2].ents)+len(n.childs[3].ents) <= maxCellEntity[N, E]Count {
	// 	n.merge()
	// }

	n.mergeFromDeep(maxCellEntityCount)

	if n.parent != nil {
		n.parent.mergeRecursively(maxCellEntityCount)
	}
}

func (n *QuadNode[N, E]) mergeFromDeep(maxCellEntityCount int) {
	for _, child := range n.childs {
		if child != nil && !child.isLeaf() {
			child.mergeFromDeep(maxCellEntityCount)
		}
	}

	if n.numSubleafs() == 4 {
		if len(n.childs[0].ents)+len(n.childs[1].ents)+len(n.childs[2].ents)+len(n.childs[3].ents) <= maxCellEntityCount {
			n.merge()
		}
		return
	}
}

func (n *QuadNode[N, E]) merge() {
	n.ents = slices.Grow(n.childs[0].ents, len(n.childs[1].ents)+len(n.childs[2].ents)+len(n.childs[3].ents))
	for i := 1; i < 4; i++ {
		n.ents = append(n.ents, n.childs[i].ents...)
	}
	for _, ent := range n.ents {
		ent.node = n
	}

	n.childs[0] = nil
	n.childs[1] = nil
	n.childs[2] = nil
	n.childs[3] = nil
}

func (n *QuadNode[N, E]) calcMid() {
	n.midX = n.minX + (n.maxX-n.minX)*0.5
	n.midY = n.minY + (n.maxY-n.minY)*0.5
}

func (n *QuadNode[N, E]) findDeepestNodeAt(x, y float64) *QuadNode[N, E] {
	if n.isLeaf() {
		return n
	}

	if y < n.midY {
		if x < n.midX {
			if n.childs[0] == nil {
				return n
			} else {
				return n.childs[0].findDeepestNodeAt(x, y)
			}
		} else {
			if n.childs[1] == nil {
				return n
			} else {
				return n.childs[1].findDeepestNodeAt(x, y)
			}
		}
	} else {
		if x < n.midX {
			if n.childs[2] == nil {
				return n
			} else {
				return n.childs[2].findDeepestNodeAt(x, y)
			}
		} else {
			if n.childs[3] == nil {
				return n
			} else {
				return n.childs[3].findDeepestNodeAt(x, y)
			}
		}
	}
}

func (n *QuadNode[N, E]) removeMarkedEntitiesRecursively() {
	if n.ents != nil {
		n.ents = slices.DeleteFunc(n.ents, func(e *Entity[N, E]) bool {
			return e == nil
		})
	} else {
		for _, child := range n.childs {
			if child != nil {
				child.removeMarkedEntitiesRecursively()
			}
		}
	}
}

type Entity[N any, E any] struct {
	x, y     float64
	node     *QuadNode[N, E]
	userData E
}

func (e *Entity[N, E]) UserData() *E {
	return &e.userData
}

func (e *Entity[N, E]) Position() (x, y float64) {
	return e.x, e.y
}

func (e *Entity[N, E]) ContainingNode() *QuadNode[N, E] {
	return e.node
}
