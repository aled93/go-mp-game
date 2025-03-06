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

package instances

import (
	"gomp"
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/systems"
	"gomp/stdsystems"
)

func NewSystemList() SystemList {
	newSystemList := SystemList{
		Player:                   systems.NewPlayerSystem(),
		Debug:                    stdsystems.NewDebugSystem(),
		Velocity:                 stdsystems.NewVelocitySystem(),
		Network:                  stdsystems.NewNetworkSystem(),
		NetworkReceive:           stdsystems.NewNetworkReceiveSystem(),
		NetworkSend:              stdsystems.NewNetworkSendSystem(),
		AnimationSpriteMatrix:    stdsystems.NewAnimationSpriteMatrixSystem(),
		AnimationPlayer:          stdsystems.NewAnimationPlayerSystem(),
		TextureRenderSprite:      stdsystems.NewTextureRenderSpriteSystem(),
		TextureRenderSpriteSheet: stdsystems.NewTextureRenderSpriteSheetSystem(),
		SpriteMatrix:             stdsystems.NewSpriteMatrixSystem(),
		AssetLib:                 stdsystems.NewAssetLibSystem([]gomp.AnyAssetLibrary{assets.Textures}),
		YSort:                    stdsystems.NewYSortSystem(),
		Collision:                stdsystems.NewCollisionSystem(),
		SpatialCollision:         stdsystems.NewSpatialCollisionSystem(),
		CollisionHandler:         systems.NewCollisionHandlerSystem(),
		Render:                   stdsystems.NewRenderSystem(),
	}

	return newSystemList
}

type SystemList struct {
	Player                   systems.PlayerSystem
	Debug                    stdsystems.DebugSystem
	Velocity                 stdsystems.VelocitySystem
	Network                  stdsystems.NetworkSystem
	NetworkReceive           stdsystems.NetworkReceiveSystem
	NetworkSend              stdsystems.NetworkSendSystem
	AnimationSpriteMatrix    stdsystems.AnimationSpriteMatrixSystem
	AnimationPlayer          stdsystems.AnimationPlayerSystem
	TextureRenderSprite      stdsystems.TextureRenderSpriteSystem
	TextureRenderSpriteSheet stdsystems.TextureRenderSpriteSheetSystem
	SpriteMatrix             stdsystems.SpriteMatrixSystem
	AssetLib                 stdsystems.AssetLibSystem
	YSort                    stdsystems.YSortSystem
	Collision                stdsystems.CollisionSystem
	CollisionHandler         systems.CollisionHandlerSystem
	SpatialCollision         stdsystems.SpatialCollisionSystem
	Render                   stdsystems.RenderSystem
}
