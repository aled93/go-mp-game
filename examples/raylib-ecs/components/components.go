/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package components

import (
	"gomp/pkg/ecs"
)

// ============
// Business
// ============

var HealthService = ecs.CreateComponentService[Health](HEALTH_ID)

// ============
// Default
// ============

var PositionService = ecs.CreateComponentService[Position](POSITION_ID)
var RotationService = ecs.CreateComponentService[Rotation](ROTATION_ID)
var ScaleService = ecs.CreateComponentService[Scale](SCALE_ID)
var MirroredService = ecs.CreateComponentService[Mirrored](MIRRORED_ID)
var InputIntentService = ecs.CreateComponentService[InputIntent](INPUT_INTENT_ID)
var LocalInputService = ecs.CreateComponentService[LocalInput](LOCAL_INPUT_ID)
var BotRoamerService = ecs.CreateComponentService[BotRoamer](BOT_ROAMER_ID)

// Rendering

var SpriteService = ecs.CreateComponentService[Sprite](SPRITE_ID)
var SpriteSheetService = ecs.CreateComponentService[SpriteSheet](SPRITE_SHEET_ID)
var SpriteMatrixService = ecs.CreateComponentService[SpriteMatrix](SPRITE_MATRIX_ID)
var TintService = ecs.CreateComponentService[Tint](TINT_ID)

var AnimationPlayerService = ecs.CreateComponentService[AnimationPlayer](ANIMATION_ID)
var AnimationStateService = ecs.CreateComponentService[AnimationState](ANIMATION_STATE_ID)

var TextureRenderService = ecs.CreateComponentService[TextureRender](TEXTURE_RENDER_ID)

// Network

var NetworkComponentService = ecs.CreateComponentService[Network](NETWORK_ID)
