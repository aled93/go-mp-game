/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package systems

import (
	"fmt"
	"gomp_game/pkgs/gomp/ecs"
	"io"
	"os"
	"runtime/pprof"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type debugController struct {
	pprofActive bool
	pprofFile   io.WriteCloser
	pprofIndex  int
}

func (s *debugController) Init(world *ecs.World) {}

func (s *debugController) Update(world *ecs.World) {
	if rl.IsKeyPressed(rl.KeyF9) {
		if s.pprofActive {
			pprof.StopCPUProfile()
			s.pprofFile.Close()
		} else {
			var err error
			s.pprofFile, err = os.Create(fmt.Sprintf("profile%d.pprof", s.pprofIndex))
			if err != nil {
				println("Failed start cpu profile:" + err.Error())
			} else {
				pprof.StartCPUProfile(s.pprofFile)
			}
		}
	}
}

func (s *debugController) FixedUpdate(world *ecs.World) {}
func (s *debugController) Destroy(world *ecs.World)     {}
