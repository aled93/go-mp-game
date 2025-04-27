package draw

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type drawCommandKind byte

const (
	drawCmd_Noop drawCommandKind = iota
	// i0: width, i1: height
	drawCmd_InitWindow
	drawCmd_DestroyWindow
	// clr: clear color
	drawCmd_Clear
	// cam2d: camera
	drawCmd_BeginMode2D
	drawCmd_EndMode2D
	// rt: render target
	drawCmd_BeginTextureMode
	drawCmd_EndTextureMode
	// i0: blend mode
	drawCmd_BeginBlendMode
	drawCmd_EndBlendMode
	// i0: width, i1: height, u0: fboDefferedId, u1: color texDefferedId, u2: depth texDefferedId
	drawCmd_CreateRenderTexture
	// u0: rtDefferedId, u1: color texDefferedId, u2: depth texDefferedId
	drawCmd_DestroyRenderTexture
	// str: filepath, u0: texDefferedId
	drawCmd_CreateTextureFromFile
	// bytes: memory, u0: texDefferedId
	drawCmd_CreateTextureFromImage
	// u0: texDefferedId
	drawCmd_DestroyTexture
	// i0: srcX, i1: srcY, i2: dstX, i3: dstY, clr: color
	drawCmd_Line
	// i0: posX, i1: posY, i2: width, i3: height, clr: color
	drawCmd_RectLine
	// i0: posX, i1: posY, i2: width, i3: height, clr: color
	drawCmd_RectFill
	// i0: posX, i1: posY, i2: width, i3: height, f0: rotation, clr: color
	drawCmd_RectFillRot
	// i0: posX, i1: posY, f0: radius, clr: color
	drawCmd_CircleFill
	// i0: posX, i1: posY, i2: fontSize, clr: color, str: text
	drawCmd_Text
	// f0-f3: srcRect, f4-f7: dstRect, f8: rotation, str: texture, clr: tint
	drawCmd_Texture
)

type drawCommand struct {
	kind drawCommandKind

	i0, i1, i2, i3 int32
	f0, f1, f2, f3 float32
	f4, f5, f6, f7 float32
	f8, f9, fA     float32

	// storing those fields for every command is wasteful, but golang doesn't
	// have unions or rust enum structs, so we pay with a lot mem usage :)
	// it is possible to store fields of those structs in intParams and
	// floatParams, but since this package is proof of concept - puck it

	clr   color.RGBA
	str   string
	tex   rl.Texture2D
	rt    rl.RenderTexture2D
	cam2d rl.Camera2D
	img   *rl.Image
}
