/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package stdsystems

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"os"
	"runtime/pprof"
)

func NewDebugSystem() DebugSystem {
	return DebugSystem{}
}

type DebugSystem struct {
	pprofEnabled bool
}

func (s *DebugSystem) Init() {}
func (s *DebugSystem) Run() {
	if rl.IsKeyPressed(rl.KeyF9) {
		if s.pprofEnabled {
			pprof.StopCPUProfile()
			fmt.Println("CPU Profile Stopped")

			// Create a memory profile file
			memProfileFile, err := os.Create("mem.out")
			if err != nil {
				panic(err)
			}
			defer memProfileFile.Close()

			// Write memory profile to file
			if err := pprof.WriteHeapProfile(memProfileFile); err != nil {
				panic(err)
			}
			fmt.Println("Memory profile written to mem.prof")
		} else {
			f, err := os.Create("cpu.out")
			if err != nil {
				log.Fatal(err)
			}

			err = pprof.StartCPUProfile(f)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("CPU Profile Started")
		}

		s.pprofEnabled = !s.pprofEnabled
	}
}
func (s *DebugSystem) Destroy() {}
