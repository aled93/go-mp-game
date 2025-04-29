package draw

import (
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Job struct {
	commands    []drawCommand
	drawEndChan chan<- time.Time
	// some commands better be called synchronously to avoid complicating api
	immediateCommand    drawCommand
	immediateResultChan chan drawCommand
}

func (job *Job) Execute() {
	if job.immediateCommand.kind != drawCmd_Noop {
		switch job.immediateCommand.kind {
		case drawCmd_InitWindow:
			rl.InitWindow(job.immediateCommand.i0, job.immediateCommand.i1, job.immediateCommand.str)
			job.immediateResultChan <- drawCommand{}

		case drawCmd_DestroyWindow:
			rl.CloseWindow()
			job.immediateResultChan <- drawCommand{}

		case drawCmd_CreateRenderTexture:
			job.immediateResultChan <- drawCommand{
				rt: rl.LoadRenderTexture(job.immediateCommand.i0, job.immediateCommand.i1),
			}

		case drawCmd_DestroyRenderTexture:
			rl.UnloadRenderTexture(job.immediateCommand.rt)
			job.immediateResultChan <- drawCommand{}

		case drawCmd_CreateTextureFromFile:
			job.immediateResultChan <- drawCommand{
				tex: rl.LoadTexture(job.immediateCommand.str),
			}

		case drawCmd_CreateTextureFromImage:
			job.immediateResultChan <- drawCommand{
				tex: rl.LoadTextureFromImage(job.immediateCommand.img),
			}

		case drawCmd_DestroyTexture:
			rl.UnloadTexture(job.immediateCommand.tex)
			job.immediateResultChan <- drawCommand{}

		default:
			log.Printf("draw command %v can't be called synchronously", job.immediateCommand.kind)
		}

		return
	}

	rl.BeginDrawing()

	for i := range job.commands {
		cmd := &job.commands[i]

		switch cmd.kind {
		case drawCmd_Clear:
			rl.ClearBackground(cmd.clr)

		case drawCmd_BeginMode2D:
			rl.BeginMode2D(cmd.cam2d)

		case drawCmd_EndMode2D:
			rl.EndMode2D()

		case drawCmd_BeginTextureMode:
			rl.BeginTextureMode(cmd.rt)

		case drawCmd_EndTextureMode:
			rl.EndTextureMode()

		case drawCmd_BeginBlendMode:
			rl.BeginBlendMode(rl.BlendMode(cmd.i0))

		case drawCmd_EndBlendMode:
			rl.EndBlendMode()

		case drawCmd_Line:
			rl.DrawLineEx(rl.NewVector2(cmd.f0, cmd.f1), rl.NewVector2(cmd.f2, cmd.f3), cmd.f4, cmd.clr)

		case drawCmd_RectLine:
			rl.DrawRectangleLinesEx(rl.NewRectangle(cmd.f0, cmd.f1, cmd.f2, cmd.f3), cmd.f4, cmd.clr)

		case drawCmd_RectFill:
			rl.DrawRectangleRec(rl.NewRectangle(cmd.f0, cmd.f1, cmd.f2, cmd.f3), cmd.clr)

		case drawCmd_RectFillRot:
			rl.DrawRectanglePro(
				rl.NewRectangle(cmd.f0, cmd.f1, cmd.f2, cmd.f3),
				rl.NewVector2(cmd.f4, cmd.f5), cmd.f6, cmd.clr,
			)

		case drawCmd_CircleFill:
			rl.DrawCircleV(rl.NewVector2(cmd.f0, cmd.f1), cmd.f3, cmd.clr)

		case drawCmd_Text:
			rl.DrawTextEx(rl.GetFontDefault(), cmd.str, rl.NewVector2(cmd.f0, cmd.f1), cmd.f2, cmd.f3, cmd.clr)

		case drawCmd_Texture:
			rl.DrawTexturePro(cmd.tex,
				rl.NewRectangle(cmd.f0, cmd.f1, cmd.f2, cmd.f3),
				rl.NewRectangle(cmd.f4, cmd.f5, cmd.f6, cmd.f7),
				rl.NewVector2(cmd.f9, cmd.fA), cmd.f8, cmd.clr,
			)
		}
	}

	rl.DrawRenderBatchActive()
	rl.SwapScreenBuffer()

	job.drawEndChan <- time.Now()
}
