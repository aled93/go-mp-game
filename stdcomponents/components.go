/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdcomponents

import (
	"gomp/pkg/ecs"
)

const (
	INVALID_ID ecs.ComponentID = iota + 128
	COLOR_ID
	POSITION_ID
	ROTATION_ID
	SCALE_ID
	FLIP_ID
	HEALTH_ID
	DESTROY_ID
	SPRITE_ID
	SPRITE_SHEET_ID
	SPRITE_MATRIX_ID
	TEXTURE_RENDER_ID
	ANIMATION_ID
	ANIMATION_PLAYER_ID
	ANIMATION_STATE_ID
	TINT_ID
	SPRITE_ANIMATOR_ID
	NETWORK_ID
)
