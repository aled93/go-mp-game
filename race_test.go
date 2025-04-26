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

package gomp

import (
	"runtime"
	"sync"
	"testing"
)

func TestRaceCondition(t *testing.T) {
	var (
		numGoroutines = runtime.NumCPU() // Количество горутин
		sliceSize     = runtime.NumCPU() // Размер слайса (меньше numGoroutines)
	)
	slice := make([]float32, sliceSize)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			index := id % sliceSize // Индекс может повторяться
			slice[index] = float32(id)
		}(i)
	}

	wg.Wait()
}

// СТЕНА ПОЗОРА MaxHero90@twitch - 15 лет опыта работы гофером и все псу под нос
func TestRaceConditionMaxHero90(t *testing.T) {
	var (
		numGoroutines = 2  // Количество горутин
		sliceSize     = 16 // Размер слайса (меньше numGoroutines)
	)
	slice := make([]float32, sliceSize)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := id; j < len(slice); j += 2 {
				slice[j] = float32(id)
			}
		}(i)
	}

	wg.Wait()
	t.Log(slice)
}
