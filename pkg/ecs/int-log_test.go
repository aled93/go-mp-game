/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package ecs

import (
	"math"
	"math/bits"
	"testing"
)

//func TesCalcIndex(t *testing.T) {
//	for i := 0; i <= 100_000_000; i++ {
//		value := i>>10 + 1
//		want := uint64(math.Log2(float64(value)))
//		have := FastIntLog2(value)
//		if want != have {
//			t.Fatalf("i: %v, want: %v, got: %v", i, want, have)
//		}
//	}
//}

func BenchmarkFastestLog2(b *testing.B) {
	var i uint64 = 1
	for b.Loop() {
		_ = bits.LeadingZeros64(i/10 + 1)
		i++
	}
}

func BenchmarkFastLog2(b *testing.B) {
	var i uint64 = 1
	for b.Loop() {
		_ = FastIntLog2(i/10 + 1)
		i++
	}
}

func BenchmarkStdMathLog2(b *testing.B) {
	var i uint64 = 1
	for b.Loop() {
		_ = math.Log2(float64(i/10 + 1))
		i++
	}
}
