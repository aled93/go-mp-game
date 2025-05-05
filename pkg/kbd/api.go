// This package mostly for debug purposes, actual game logic code
// should use something like input actions which can be rebinded,
// can be triggered by any input device and so on
package kbd

var (
	downKeycodes      map[Keycode]bool
	prevDownKeycodes  map[Keycode]bool
	downScancodes     map[Scancode]bool
	prevDownScancodes map[Scancode]bool
)

func IsKeyDown(keycode Keycode) bool {
	down := downKeycodes[keycode]
	return down
}

func IsKeyPressed(keycode Keycode) bool {
	down := downKeycodes[keycode]
	prevDown := prevDownKeycodes[keycode]
	return down && !prevDown
}

func IsKeyReleased(keycode Keycode) bool {
	down := downKeycodes[keycode]
	prevDown := prevDownKeycodes[keycode]
	return !down && prevDown
}

func IsScancodeDown(scancode Scancode) bool {
	down := downScancodes[scancode]
	return down
}

func IsScancodePressed(scancode Scancode) bool {
	down := downScancodes[scancode]
	prevDown := prevDownScancodes[scancode]
	return down && !prevDown
}

func IsScancodeReleased(scancode Scancode) bool {
	down := downScancodes[scancode]
	prevDown := prevDownScancodes[scancode]
	return !down && prevDown
}

func SetDown(keycode Keycode, scancode Scancode) {
	downKeycodes[keycode] = true
}

func SetUp(keycode Keycode, scancode Scancode) {
	downKeycodes[keycode] = false
}

func Update() {
	downKeycodes, prevDownKeycodes = prevDownKeycodes, downKeycodes
	downScancodes, prevDownScancodes = prevDownScancodes, downScancodes
	clear(downKeycodes)
	clear(downScancodes)
}
