// This package mostly for debug purposes, actual game logic code
// should use something like input actions which can be rebinded,
// can be triggered by any input device and so on
package kbd

var (
	downKeycodes     map[Keycode]bool
	prevDownKeycodes map[Keycode]bool
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

func SetDown(keycode Keycode, scancode Scancode) {
	downKeycodes[keycode] = true
}

func SetUp(keycode Keycode, scancode Scancode) {
	downKeycodes[keycode] = false
}

func Update() {
	downKeycodes, prevDownKeycodes = prevDownKeycodes, downKeycodes
	clear(downKeycodes)
}
