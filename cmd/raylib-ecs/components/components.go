/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package components

import (
	"gomp_game/cmd/raylib-ecs/gravity"
	"gomp_game/pkgs/gomp/ecs"
	"gomp_game/pkgs/spatial"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Position struct {
	X, Y float32
}
type Rotation struct {
	Angle float32
}
type Scale struct {
	X, Y float32
}

type Health struct {
	Hp, MaxHp int32
}

type Sprite struct {
	Texture       *rl.Texture2D
	TextureRegion rl.Rectangle
	Origin        rl.Vector2
	Tint          color.RGBA
}

type Velocity struct {
	X, Y float32
}

type SpatialEnt struct {
	Ent *spatial.Entity[gravity.QuadNodeUserData, any]
}

type GravitationEmitter struct{}

type GravitationReceiver struct{}

type SpriteRender struct {
	Sprite   Sprite
	Dest     rl.Rectangle
	Rotation float32
}

var PositionService = ecs.CreateComponentService[Position](TRANSFORM_ID)
var RotationService = ecs.CreateComponentService[Rotation](ROTATION_ID)
var ScaleService = ecs.CreateComponentService[Scale](SCALE_ID)
var HealthService = ecs.CreateComponentService[Health](HEALTH_ID)
var SpriteService = ecs.CreateComponentService[Sprite](SPRITE_ID)
var VelocityService = ecs.CreateComponentService[Velocity](VELOCITY_ID)
var SpatialEntService = ecs.CreateComponentService[SpatialEnt](SPATIALENT_ID)
var GravitationEmitterService = ecs.CreateComponentService[GravitationEmitter](GRAVEMITTER_ID)
var GravitationReceiveService = ecs.CreateComponentService[GravitationReceiver](GRAVRECEIVER_ID)
var SpriteRenderService = ecs.CreateComponentService[SpriteRender](SPRITE_RENDER_ID)

// spawn creature every tick with random hp and position
// each creature looses hp every tick
// each creature has Color that depends on its own maxHP and current hp
// when hp == 0 creature dies

// spawn system
// movement system
// hp system
// Destroy system
