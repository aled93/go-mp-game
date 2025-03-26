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

package systems

import (
	"gomp/examples/new-api/assets"
	"gomp/examples/new-api/components"
	"gomp/pkg/ecs"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func NewAudioSystem() AudioSystem {
	return AudioSystem{}
}

type AudioSystem struct {
	EntityManager    *ecs.EntityManager
	SpaceshipIntents *components.SpaceshipIntentComponentManager
}

func (s *AudioSystem) Init() {
	rl.InitAudioDevice()

	assets.Audio.Load("damage_sound.wav")
	assets.Audio.Load("fly_sound.wav")
	assets.Audio.Load("gun_sound.wav")
}
func (s *AudioSystem) Run(dt time.Duration) {}
func (s *AudioSystem) Destroy() {
	rl.CloseAudioDevice()
}
