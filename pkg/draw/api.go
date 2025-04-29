package draw

import (
	"image/color"
	"sync/atomic"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/negrel/assert"
)

var (
	activeJob *Job = &Job{
		commands:    make([]drawCommand, 0, 0xFFFF),
		drawEndChan: frameEndTimeChan,
	}
	nextJob *Job = &Job{
		commands:    make([]drawCommand, 0, 0xFFFF),
		drawEndChan: frameEndTimeChan,
	}
	jobProcessor chan<- Job

	frameEndTimeChan      = make(chan time.Time, 1)
	frameEndFirst         time.Time
	frameDrawnNum         int
	fpsSmoothed           atomic.Int64
	fpsSmoothedLastUpdate time.Time
)

func init() {
	go func() {
		for frameEnd := range frameEndTimeChan {
			if frameDrawnNum == 0 {
				frameEndFirst = frameEnd
			}
			frameDrawnNum++

			// update every second
			if time.Since(fpsSmoothedLastUpdate) < time.Second {
				continue
			}
			fpsSmoothedLastUpdate = time.Now()

			timeRange := frameEnd.Sub(frameEndFirst)
			fps := int64(float64(frameDrawnNum) / timeRange.Seconds())
			fpsSmoothed.Store(fps)

			frameDrawnNum = 0
		}
	}()
}

func runCommandSync(cmd drawCommand) drawCommand {
	resultChan := make(chan drawCommand)
	defer close(resultChan)

	jobProcessor <- Job{
		immediateResultChan: resultChan,
		immediateCommand:    cmd,
	}

	return <-resultChan
}

// SetJobProcessor sets channel which one will be sent to draw jobs.
// Overrides any previous processor.
func SetJobProcessor(processor chan<- Job) {
	jobProcessor = processor
}

func BeginDrawing() {
	if activeJob.commands != nil {
		clear(activeJob.commands)
		activeJob.commands = activeJob.commands[:0]
	}
}

func EndDrawing() {
	assert.NotNil(jobProcessor, "draw job processor is nil")

	jobProcessor <- *activeJob
	activeJob, nextJob = nextJob, activeJob
}

func InitWindow(width, height int32, title string) {
	runCommandSync(drawCommand{
		kind: drawCmd_InitWindow,
		i0:   width,
		i1:   height,
		str:  title,
	})
}

func DestroyWindow() {
	runCommandSync(drawCommand{
		kind: drawCmd_DestroyWindow,
	})
}

func GetFPS() int64 {
	return fpsSmoothed.Load()
}

func ClearBackground(color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_Clear,
		clr:  color,
	})
}

func BeginMode2D(cam rl.Camera2D) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind:  drawCmd_BeginMode2D,
		cam2d: cam,
	})
}

func EndMode2D() {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_EndMode2D,
	})
}

func BeginTextureMode(rt rl.RenderTexture2D) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_BeginTextureMode,
		rt:   rt,
	})
}

func EndTextureMode() {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_EndTextureMode,
	})
}

func BeginBlendMode(blend rl.BlendMode) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_BeginBlendMode,
		i0:   int32(blend),
	})
}

func EndBlendMode() {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_EndBlendMode,
	})
}

func CreateRenderTexture(width, height int32) (rt rl.RenderTexture2D) {
	return runCommandSync(drawCommand{
		kind: drawCmd_CreateRenderTexture,
		i0:   width,
		i1:   height,
	}).rt
}

func DestroyRenderTexture(rt rl.RenderTexture2D) {
	runCommandSync(drawCommand{
		kind: drawCmd_DestroyRenderTexture,
		rt:   rt,
	})
}

func CreateTextureFromFile(filePath string) (tex rl.Texture2D) {
	return runCommandSync(drawCommand{
		kind: drawCmd_CreateTextureFromFile,
		str:  filePath,
	}).tex
}

func CreateTextureFromImage(image *rl.Image) (tex rl.Texture2D) {
	return runCommandSync(drawCommand{
		kind: drawCmd_CreateTextureFromImage,
		img:  image,
	}).tex
}

func DestroyTexture(tex rl.Texture2D) {
	runCommandSync(drawCommand{
		kind: drawCmd_DestroyTexture,
		tex:  tex,
	})
}

func Line(x0, y0, x1, y1, thickness float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_Line,
		f0:   x0,
		f1:   y0,
		f2:   x1,
		f3:   y1,
		f4:   thickness,
		clr:  color,
	})
}

func RectLine(x, y, w, h, thickness float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_RectLine,
		f0:   x,
		f1:   y,
		f2:   w,
		f3:   h,
		f4:   thickness,
		clr:  color,
	})
}

func RectFill(x, y, w, h float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_RectFill,
		f0:   x,
		f1:   y,
		f2:   w,
		f3:   h,
		clr:  color,
	})
}

func RectFillAngled(x, y, w, h float32, ang float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_RectFillRot,
		f0:   x,
		f1:   y,
		f2:   w,
		f3:   h,
		f4:   ang,
		clr:  color,
	})
}

func CircleFill(x, y, r float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_CircleFill,
		f0:   x,
		f1:   y,
		f2:   r,
		clr:  color,
	})
}

func Text(text string, x, y, fontSize, spacing float32, color color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_Text,
		f0:   x,
		f1:   y,
		f2:   fontSize,
		f3:   spacing,
		clr:  color,
		str:  text,
	})
}

func Texture(texture rl.Texture2D, src rl.Rectangle, dst rl.Rectangle, orig rl.Vector2, rot float32, tint color.RGBA) {
	assert.NotNil(activeJob, "drawing didn't started")

	activeJob.commands = append(activeJob.commands, drawCommand{
		kind: drawCmd_Texture,
		f0:   src.X,
		f1:   src.Y,
		f2:   src.Width,
		f3:   src.Height,
		f4:   dst.X,
		f5:   dst.Y,
		f6:   dst.Width,
		f7:   dst.Height,
		f8:   rot,
		f9:   orig.X,
		fA:   orig.Y,
		tex:  texture,
		clr:  tint,
	})
}
