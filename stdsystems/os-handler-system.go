/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"gomp/pkg/ecs"
	"gomp/pkg/kbd"
	"gomp/stdcomponents"
	"time"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

func NewOSHandlerSystem() OSHandlerSystem {
	return OSHandlerSystem{}
}

type OSHandlerSystem struct {
	Entities *ecs.EntityManager
	Windows  *stdcomponents.WindowComponentManager

	mainWindowEntId ecs.Entity
}

func (s *OSHandlerSystem) Init(entities *ecs.EntityManager) {
	s.mainWindowEntId = entities.Create()

	s.Windows.Create(s.mainWindowEntId, stdcomponents.Window{
		Handle: sdl.CreateWindow("GOMP", 1280, 720, sdl.WindowFlags(0)),
	})
}

func (s *OSHandlerSystem) Run(dt time.Duration) bool {
	kbd.Update()

	var ev sdl.Event
	for sdl.PollEvent(&ev) {
		switch ev.Type() {
		case sdl.EventAudioDeviceAdded:
		case sdl.EventAudioDeviceFormatChanged:
		case sdl.EventAudioDeviceRemoved:
		case sdl.EventCameraDeviceAdded:
		case sdl.EventCameraDeviceApproved:
		case sdl.EventCameraDeviceDenied:
		case sdl.EventCameraDeviceRemoved:
		case sdl.EventClipboardUpdate:
		case sdl.EventDidEnterBackground:
		case sdl.EventDidEnterForeground:
		case sdl.EventDisplayAdded:
		case sdl.EventDisplayContentScaleChanged:
		case sdl.EventDisplayCurrentModeChanged:
		case sdl.EventDisplayDesktopModeChanged:
		case sdl.EventDisplayMoved:
		case sdl.EventDisplayOrientation:
		case sdl.EventDisplayRemoved:
		case sdl.EventDropBegin:
		case sdl.EventDropComplete:
		case sdl.EventDropFile:
		case sdl.EventDropPosition:
		case sdl.EventDropText:
		case sdl.EventEnumPadding:
		case sdl.EventFingerCanceled:
		case sdl.EventFingerDown:
		case sdl.EventFingerMotion:
		case sdl.EventFingerUp:
		case sdl.EventGamepadAdded:
		case sdl.EventGamepadAxisMotion:
		case sdl.EventGamepadButtonDown:
		case sdl.EventGamepadButtonUp:
		case sdl.EventGamepadRemapped:
		case sdl.EventGamepadRemoved:
		case sdl.EventGamepadSensorUpdate:
		case sdl.EventGamepadSteamHandleUpdated:
		case sdl.EventGamepadTouchpadDown:
		case sdl.EventGamepadTouchpadMotion:
		case sdl.EventGamepadTouchpadUp:
		case sdl.EventGamepadUpdateComplete:
		case sdl.EventJoystickAdded:
		case sdl.EventJoystickAxisMotion:
		case sdl.EventJoystickBallMotion:
		case sdl.EventJoystickBatteryUpdated:
		case sdl.EventJoystickButtonDown:
		case sdl.EventJoystickButtonUp:
		case sdl.EventJoystickHatMotion:
		case sdl.EventJoystickRemoved:
		case sdl.EventJoystickUpdateComplete:
		case sdl.EventKeyDown:
			ev := ev.Key()
			switch ev.Key {
			case sdl.KeycodeEscape:
				return true
			}

			kbd.SetDown(kbd.Keycode(ev.Key), kbd.Scancode(ev.Scancode))

		case sdl.EventKeyUp:
			ev := ev.Key()
			kbd.SetUp(kbd.Keycode(ev.Key), kbd.Scancode(ev.Scancode))

		case sdl.EventKeyboardAdded:
		case sdl.EventKeyboardRemoved:
		case sdl.EventKeymapChanged:
		case sdl.EventLocaleChanged:
		case sdl.EventLowMemory:
		case sdl.EventMouseAdded:
		case sdl.EventMouseButtonDown:
		case sdl.EventMouseButtonUp:
		case sdl.EventMouseMotion:
		case sdl.EventMouseRemoved:
		case sdl.EventMouseWheel:
		case sdl.EventPenAxis:
		case sdl.EventPenButtonDown:
		case sdl.EventPenButtonUp:
		case sdl.EventPenDown:
		case sdl.EventPenMotion:
		case sdl.EventPenProximityIn:
		case sdl.EventPenProximityOut:
		case sdl.EventPenUp:
		case sdl.EventPollSentinel:
		case sdl.EventQuit:
			return true

		case sdl.EventRenderDeviceLost:
		case sdl.EventRenderDeviceReset:
		case sdl.EventRenderTargetsReset:
		case sdl.EventSensorUpdate:
		case sdl.EventSystemThemeChanged:
		case sdl.EventTerminating:
		case sdl.EventTextEditing:
		case sdl.EventTextEditingCandidates:
		case sdl.EventTextInput:
		case sdl.EventUser:
		case sdl.EventWillEnterBackground:
		case sdl.EventWillEnterForeground:
		case sdl.EventWindowCloseRequested:
		case sdl.EventWindowDestroyed:
		case sdl.EventWindowDisplayChanged:
		case sdl.EventWindowDisplayScaleChanged:
		case sdl.EventWindowEnterFullscreen:
		case sdl.EventWindowExposed:
		case sdl.EventWindowFocusGained:
		case sdl.EventWindowFocusLost:
		case sdl.EventWindowHdrStateChanged:
		case sdl.EventWindowHidden:
		case sdl.EventWindowHitTest:
		case sdl.EventWindowIccprofChanged:
		case sdl.EventWindowLeaveFullscreen:
		case sdl.EventWindowMaximized:
		case sdl.EventWindowMetalViewResized:
		case sdl.EventWindowMinimized:
		case sdl.EventWindowMouseEnter:
		case sdl.EventWindowMouseLeave:
		case sdl.EventWindowMoved:
		case sdl.EventWindowOccluded:
		case sdl.EventWindowPixelSizeChanged:
		case sdl.EventWindowResized:
		case sdl.EventWindowRestored:
		case sdl.EventWindowSafeAreaChanged:
		case sdl.EventWindowShown:
		}
	}

	return false
}

func (s *OSHandlerSystem) Destroy() {
	wnd := s.Windows.GetUnsafe(s.mainWindowEntId)
	sdl.DestroyWindow(wnd.Handle)

	s.Entities.Delete(s.mainWindowEntId)
}
