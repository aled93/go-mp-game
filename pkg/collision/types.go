package collision

import (
	"fmt"
	"gomp/pkg/debugdraw"
	"gomp/pkg/spatial"
	"math"
	"time"

	"github.com/negrel/assert"
)

type Space2D struct {
	partitioner    spatial.Partitioner2D[*Object]
	maxObjHalfSize float64
}

type SpaceCreateParams struct {
	// Choose space partition strategy based on your game, in short:
	// games with levels like room\arena with transitions between them should
	// choose implementations based on spatial.FinitePartitioner2D,
	// games with open world should use spatial.Partitioner2D based
	// implementations.
	Partitioner spatial.Partitioner2D[*Object]

	// Objects with a size greater that MaxObjectSize cannot be created.
	// Must be positive.
	MaxObjectSize float64
}

func NewSpace(params *SpaceCreateParams) *Space2D {
	assert.NotNil(params, "space creation params must be provided")
	assert.NotNil(params.Partitioner, "Partitioner must be provided")
	assert.Greater(params.MaxObjectSize, 0.0, "MaxObjectSize must be positive")

	return &Space2D{
		partitioner:    params.Partitioner,
		maxObjHalfSize: params.MaxObjectSize * 0.5,
	}
}

type ObjectCreateParams struct {
	Shape         ShapeKind
	ShapeSize     float64
	CollisionMask CollisionMask
	X, Y          float64
	IsTrigger     bool
}

func (space *Space2D) CreateObject(params *ObjectCreateParams) *Object {
	assert.NotNil(params)

	obj := &Object{
		space:   space,
		shape:   params.Shape,
		size:    params.ShapeSize,
		mask:    params.CollisionMask,
		x:       params.X,
		y:       params.Y,
		trigger: params.IsTrigger,
	}
	bounds := spatial.NewAABBFromCenterSize(params.X, params.Y, params.ShapeSize*2, params.ShapeSize*2)

	obj.spatialObj = space.partitioner.Register(obj, bounds)
	return obj
}

func (space *Space2D) DebugDraw(duration time.Duration) {
	grid, ok := space.partitioner.(*spatial.BoundlessGrid2D[*Object])
	if !ok {
		return
	}

	grid.DebugDraw(duration)
}

type Object struct {
	space      *Space2D
	spatialObj spatial.SpatialObject[*Object]
	shape      ShapeKind
	size       float64 // side half size for box, radius for circle
	mask       CollisionMask
	x, y       float64
	trigger    bool
}

func (obj *Object) spatialBounds() spatial.AABB2D {
	return spatial.NewAABBFromCenterSize(obj.x, obj.y, obj.size*2, obj.size*2)
}

func (obj *Object) Pos() (X float64, Y float64) {
	return obj.x, obj.y
}

func (obj *Object) Size() float64 {
	return obj.size
}

// CanCollideWith checks only collision masks and trigger field of both objects
func (obj *Object) CanCollideWith(other *Object) bool {
	return !obj.trigger && !other.trigger && (obj.mask&other.mask) != 0
}

func (obj *Object) Move(dx, dy float64) (t float64) {
	newArea := obj.spatialBounds()
	newArea.Shift(dx, dy)

	t = math.MaxFloat64
	var minOther *Object

	for other := range obj.space.partitioner.QueryAreaIter(newArea) {
		if other == obj.spatialObj || !obj.CanCollideWith(other.GetData()) {
			continue
		}
		otherObj := other.GetData()

		debugdraw.Line(float32(obj.x), float32(obj.y), 0.0, float32(otherObj.x), float32(otherObj.y), 0.0, 0.0, 0.0, 1.0, 1.0, 0)

		minXt := 1.0
		minYt := 1.0

		if dx < 0.0 {
			minXt = math.Abs((obj.x-obj.size)-(otherObj.x+otherObj.size)) / -dx
		} else if dx > 0.0 {
			minXt = math.Abs((otherObj.x-otherObj.size)-(obj.x+obj.size)) / dx
		}

		if dy < 0.0 {
			minYt = math.Abs((obj.y-obj.size)-(otherObj.y+otherObj.size)) / -dy
		} else if dy > 0.0 {
			minYt = math.Abs((otherObj.y-otherObj.size)-(obj.y+obj.size)) / dy
		}

		fmt.Printf("min xt=%6.4f | yt=%6.4f\n", minXt, minYt)

		// if minXt < 0.0 && minYt < 0.0 {
		// 	// otherObj is fully inside obj
		// 	continue
		// }

		curT := min(minXt, minYt)
		if curT >= 0.0 && curT <= 1.0 && curT < t {
			t = curT
			minOther = otherObj
		}
	}

	if minOther == nil {
		// no collisions detected
		t = 1.0
	} else {
		debugdraw.Line(float32(obj.x), float32(obj.y), 0.0, float32(minOther.x), float32(minOther.y), 0.0, 0.0, 1.0, 0.0, 1.0, 0)
	}

	obj.x += dx * t
	obj.y += dy * t

	newArea.Shift(-dx, -dy)
	newArea.Shift(dx*t, dy*t)

	obj.space.partitioner.UpdateObject(obj.spatialObj, obj, newArea)

	return t
}

// func (obj *Object) PlaceInClearSpace(padding float64) {
// 	var firstOccluder *Object
// 	for other := range obj.space.partitioner.QueryAreaIter(obj.spatialBounds()) {
// 		if other == obj.spatialObj || !obj.CanCollideWith(other.GetData()) {
// 			continue
// 		}

// 		firstOccluder = other.GetData()
// 		break
// 	}

// 	if firstOccluder == nil {
// 		return
// 	}

// 	occludedArea := firstOccluder.spatialBounds()
// 	for {
// 		newPoses := [4][2]float64{
// 			{occludedArea.X - obj.size - padding, obj.y},
// 			{occludedArea.X + occludedArea.W + obj.size + padding, obj.y},
// 			{obj.x, occludedArea.Y - obj.size - padding},
// 			{obj.x, occludedArea.Y + occludedArea.H + obj.size + padding},
// 		}

// 		slices.SortFunc(newPoses[:], func(a, b [2]float64) int {
// 			distA := (a[0]-obj.x)*(a[0]-obj.x) + (a[1]-obj.y)*(a[1]-obj.y)
// 			distB := (b[0]-obj.x)*(b[0]-obj.x) + (b[1]-obj.y)*(b[1]-obj.y)
// 			if distA > distB {
// 				return 1
// 			} else if distA < distB {
// 				return -1
// 			} else {
// 				return 0
// 			}
// 		})

// 		for _, newPos := range newPoses {
// 			newArea := spatial.Rectangle2D{
// 				X: newPos[0] - obj.size,
// 				Y: newPos[1] - obj.size,
// 				W: obj.size * 2.0,
// 				H: obj.size * 2.0,
// 			}
// 			for other := range obj.space.partitioner.QueryAreaIter(newArea) {
// 				if other == obj.spatialObj || !obj.CanCollideWith(other.GetData()) {
// 					continue
// 				}

// 				// TODO
// 			}
// 		}
// 	}
// }

type ShapeKind uint8

const (
	Box ShapeKind = 1 << iota
	// Circle
)

type CollisionMask = uint64

const (
	CollideNone CollisionMask = 0
	CollideAll  CollisionMask = 0xFFFF_FFFF_FFFF_FFFF
)
