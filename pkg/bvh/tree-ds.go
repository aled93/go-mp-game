package bvh

import (
	"gomp/pkg/ecs"
	"gomp/stdcomponents"
	"gomp/vectors"
	"math"
	"math/bits"
	"slices"
)

func NewTree2Du(layer stdcomponents.CollisionLayer, prealloc int) Tree2Du {
	return Tree2Du{
		layer:       layer,
		components:  make([]treeComponent, 0, prealloc),
		innerNodes:  make([]InnerNode, 0, prealloc),
		leafNodes:   make([]LeafNode, 0, prealloc),
		mortonCodes: make([]uint32, 0, prealloc),
		rootIndex:   -1,
	}
}

type Tree2Du struct {
	layer       stdcomponents.CollisionLayer
	innerNodes  []InnerNode // Отдельный массив для внутренних нод
	components  []treeComponent
	leafNodes   []LeafNode // Отдельный массив для листьев
	mortonCodes []uint32   // Кэшированные Morton codes для листьев
	rootIndex   int
}

type LeafNode struct {
	Entity ecs.Entity
	Bounds stdcomponents.AABB
}

type InnerNode struct {
	Left, Right int // Отрицательные индексы = leafNodes[-(index+1)]
	Bounds      stdcomponents.AABB
	SplitAxis   int // 0 = X, 1 = Y для SAH оптимизации
}

func (t *Tree2Du) AddComponent(entity ecs.Entity, aabbs *stdcomponents.AABB) {
	t.components = append(t.components, treeComponent{
		Entity: entity,
		AABB:   aabbs,
	})
}

func (t *Tree2Du) Layer() stdcomponents.CollisionLayer {
	return t.layer
}

func (t *Tree2Du) Build() {
	t.innerNodes = t.innerNodes[:0]
	t.leafNodes = t.leafNodes[:0]
	t.mortonCodes = t.mortonCodes[:0]

	if len(t.components) == 0 {
		t.rootIndex = -1
		return
	}

	// 1. Создаем листья с Morton codes
	leaves := make([]LeafNode, len(t.components))
	t.mortonCodes = make([]uint32, len(t.components))

	for i := range t.components {
		component := &t.components[i]
		aabb := component.AABB
		center := aabb.Min.Add(aabb.Max).Scale(0.5)
		leaves[i] = LeafNode{
			Entity: component.Entity,
			Bounds: *aabb,
		}
		t.mortonCodes[i] = t.morton2D(center.X, center.Y)
	}
	t.components = t.components[:0]

	// 2. Сортируем листья по Morton code
	indices := make([]int, len(leaves))
	for i := range indices {
		indices[i] = i
	}
	slices.SortFunc(indices, func(a, b int) int {
		return int(t.mortonCodes[a]) - int(t.mortonCodes[b])
	})

	// Реорганизуем данные в соответствии с сортировкой
	sortedLeaves := make([]LeafNode, len(leaves))
	sortedCodes := make([]uint32, len(leaves))
	for i, idx := range indices {
		sortedLeaves[i] = leaves[idx]
		sortedCodes[i] = t.mortonCodes[idx]
	}
	t.leafNodes = sortedLeaves
	t.mortonCodes = sortedCodes

	// 3. Строим иерархию используя бинарное разбиение
	if len(t.leafNodes) == 1 {
		t.rootIndex = -1 // Помечаем как единственный лист
		return
	}

	// Рекурсивное построение через стек задач
	type buildTask struct {
		start, end int
		parentPtr  *int
	}

	var stack []buildTask
	rootTask := buildTask{0, len(t.leafNodes) - 1, &t.rootIndex}
	stack = append(stack, rootTask)

	for len(stack) > 0 {
		task := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		start, end := task.start, task.end
		if start == end {
			*task.parentPtr = -start - 1 // Листовой узел
			continue
		}

		// Находим точку разделения по Morton codes
		split := t.findSplit(start, end)

		// Создаем внутреннюю ноду
		nodeIndex := len(t.innerNodes)
		t.innerNodes = append(t.innerNodes, InnerNode{})
		*task.parentPtr = nodeIndex

		// Обрабатываем детей
		stack = append(stack, buildTask{
			start:     split + 1,
			end:       end,
			parentPtr: &t.innerNodes[nodeIndex].Right,
		})
		stack = append(stack, buildTask{
			start:     start,
			end:       split,
			parentPtr: &t.innerNodes[nodeIndex].Left,
		})
	}

	// 4. Вычисляем AABB для всех нод
	t.computeBounds(t.rootIndex)
}

func (t *Tree2Du) findSplit(start, end int) int {
	first := t.mortonCodes[start]
	last := t.mortonCodes[end]
	if first == last {
		return (start + end) >> 1
	}

	commonPrefix := bits.LeadingZeros32(first ^ last)
	split := start
	step := end - start

	for {
		step = (step + 1) >> 1
		newSplit := split + step

		if newSplit < end {
			splitCode := t.mortonCodes[newSplit]
			splitPrefix := bits.LeadingZeros32(first ^ splitCode)
			if splitPrefix > commonPrefix {
				split = newSplit
			}
		}

		if step <= 1 {
			break
		}
	}

	return split
}

func (t *Tree2Du) computeBounds(nodeIndex int) stdcomponents.AABB {
	if nodeIndex < 0 { // Лист
		leaf := t.leafNodes[-nodeIndex-1]
		return leaf.Bounds
	}

	node := &t.innerNodes[nodeIndex]
	leftBounds := t.computeBounds(node.Left)
	rightBounds := t.computeBounds(node.Right)
	node.Bounds = t.mergeAABB(&leftBounds, &rightBounds)

	// Определяем ось разделения по Morton codes
	leftCode := t.getMortonCode(node.Left)
	rightCode := t.getMortonCode(node.Right)
	diff := leftCode ^ rightCode
	node.SplitAxis = bits.TrailingZeros32(diff) % 2

	return node.Bounds
}

func (t *Tree2Du) getMortonCode(nodeIndex int) uint32 {
	if nodeIndex < 0 {
		return t.mortonCodes[-nodeIndex-1]
	}
	// Для внутренних нод возвращаем код разделения
	return t.mortonCodes[t.findSplitForNode(nodeIndex)]
}

func (t *Tree2Du) findSplitForNode(nodeIndex int) int {
	node := t.innerNodes[nodeIndex]
	left := node.Left
	for left >= 0 {
		left = t.innerNodes[left].Left
	}
	return -left - 1
}

func (t *Tree2Du) Query(aabb *stdcomponents.AABB, result []ecs.Entity) []ecs.Entity {
	if t.rootIndex == -1 {
		if len(t.leafNodes) == 1 && t.aabbOverlap(&t.leafNodes[0].Bounds, aabb) {
			result = append(result, t.leafNodes[0].Entity)
		}
		return result
	}

	stack := make([]int, 0, 32)
	stack = append(stack, t.rootIndex)

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if nodeIndex >= 0 { // Inner node
			node := t.innerNodes[nodeIndex]

			if !t.aabbOverlap(&node.Bounds, aabb) {
				continue
			}

			// Оптимизация порядка обхода на основе оси разделения
			if (node.SplitAxis == 0 && aabb.Min.X <= node.Bounds.Min.X) ||
				(node.SplitAxis == 1 && aabb.Min.Y <= node.Bounds.Min.Y) {
				stack = append(stack, node.Right, node.Left)
			} else {
				stack = append(stack, node.Left, node.Right)
			}
		} else { // Leaf node
			leafIndex := -nodeIndex - 1
			if t.aabbOverlap(&t.leafNodes[leafIndex].Bounds, aabb) {
				result = append(result, t.leafNodes[leafIndex].Entity)
			}
		}
	}

	return result
}

// go:inline aabbOverlap checks if two AABB intersect
func (t *Tree2Du) aabbOverlap(a, b *stdcomponents.AABB) bool {
	// Check for non-overlap conditions first (early exit)
	if a.Max.X < b.Min.X || a.Min.X > b.Max.X {
		return false
	}
	if a.Max.Y < b.Min.Y || a.Min.Y > b.Max.Y {
		return false
	}
	return true
}

// Expands a 10-bit integer into 20 bits by inserting 1 zero after each bit
func (t *Tree2Du) expandBits2D(v uint32) uint32 {
	v = (v * 0x00010001) & 0xFF0000FF
	v = (v * 0x00000101) & 0x0F00F00F
	v = (v * 0x00000011) & 0xC30C30C3
	v = (v * 0x00000005) & 0x24924924
	return v
}

// 2D Morton code for centroids coordinates in [0,1] range
func (t *Tree2Du) morton2D(x, y float32) uint32 {
	xx := uint32(math.Min(math.Max(float64(x)*1024.0, 0.0), 1023.0))
	yy := uint32(math.Min(math.Max(float64(y)*1024.0, 0.0), 1023.0))
	return (t.expandBits2D(xx) << 1) | t.expandBits2D(yy)
}

// mergeAABB combines two AABB
func (t *Tree2Du) mergeAABB(a, b *stdcomponents.AABB) stdcomponents.AABB {
	return stdcomponents.AABB{
		Min: vectors.Vec2{
			X: min(a.Min.X, b.Min.X),
			Y: min(a.Min.Y, b.Min.Y),
		},
		Max: vectors.Vec2{
			X: max(a.Max.X, b.Max.X),
			Y: max(a.Max.Y, b.Max.Y),
		},
	}
}
