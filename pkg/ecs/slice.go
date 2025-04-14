/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package ecs

import (
	"sync"

	"github.com/negrel/assert"
)

type Slice[T any] struct {
	data []T
	len  int
}

func NewSlice[T any](size int) (a Slice[T]) {
	a.data = make([]T, 0, size)
	return a
}

func (a *Slice[T]) Len() int {
	return a.len
}

func (a *Slice[T]) Get(index int) *T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	return &a.data[index]
}

func (a *Slice[T]) GetValue(index int) T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	return a.data[index]
}

func (a *Slice[T]) Set(index int, value T) *T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	a.data[index] = value

	return &a.data[index]
}

func (a *Slice[T]) Append(values ...T) *T {
	a.data = append(a.data[:a.len], values...)
	a.len += len(values)
	return &a.data[a.len-1]
}

func (a *Slice[T]) SoftReduce() {
	assert.True(a.len > 0, "Len is already 0")
	a.len--
}

func (a *Slice[T]) Reset() {
	a.len = 0
}

func (a *Slice[T]) Copy(fromIndex, toIndex int) {
	assert.True(fromIndex >= 0, "index out of range")
	assert.True(fromIndex < a.len, "index out of range")
	from := a.Get(fromIndex)

	assert.True(toIndex >= 0, "index out of range")
	assert.True(toIndex < a.len, "index out of range")
	to := a.Get(toIndex)

	*to = *from
}

func (a *Slice[T]) Swap(i, j int) (newI, NewJ *T) {
	assert.True(i >= 0, "index out of range")
	assert.True(i < a.len, "index out of range")
	x := a.Get(i)

	assert.True(j >= 0, "index out of range")
	assert.True(j < a.len, "index out of range")
	y := a.Get(j)

	*x, *y = *y, *x
	return x, y
}

func (a *Slice[T]) Last() *T {
	index := a.len - 1
	assert.True(index >= 0, "index out of range")

	return a.Get(index)
}

func (a *Slice[T]) Raw() []T {
	return a.data
}

func (a *Slice[T]) getPageIdAndIndex(index int) (int, int) {
	pageId := index >> pageSizeShift
	assert.True(pageId < len(a.data), "index out of range")

	index %= pageSize
	assert.True(index < pageSize, "index out of range")

	return pageId, index
}

func (a *Slice[T]) Each(yield func(int, *T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(j, &a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) EachData(yield func(*T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(&a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) EachDataValue(yield func(T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) EachParallel(numWorkers int, yield func(int, *T, int) bool) {
	assert.True(numWorkers > 0)
	var chunkSize = a.len / numWorkers
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for workedId := 0; workedId < numWorkers; workedId++ {
		startIndex := workedId * chunkSize
		endIndex := startIndex + chunkSize - 1
		if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
			endIndex = a.len
		}
		go func(start int, end int) {
			defer wg.Done()
			for i := range a.data[start:end] {
				if !yield(i, &a.data[i+startIndex], workedId) {
					return
				}
			}
		}(startIndex, endIndex)
	}
	wg.Wait()
}

func (a *Slice[T]) EachDataValueParallel(numWorkers int, yield func(T, int) bool) {
	assert.True(numWorkers > 0)
	var chunkSize = a.len / numWorkers
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for workedId := 0; workedId < numWorkers; workedId++ {
		startIndex := workedId * chunkSize
		endIndex := startIndex + chunkSize - 1
		if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
			endIndex = a.len
		}
		go func(start int, end int) {
			defer wg.Done()
			for i := range a.data[start:end] {
				if !yield(a.data[i+startIndex], workedId) {
					return
				}
			}
		}(startIndex, endIndex)
	}
	wg.Wait()
}

func (a *Slice[T]) EachDataParallel(numWorkers int, yield func(*T, int) bool) {
	assert.True(numWorkers > 0)
	var chunkSize = a.len / numWorkers
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for workedId := 0; workedId < numWorkers; workedId++ {
		startIndex := workedId * chunkSize
		endIndex := startIndex + chunkSize - 1
		if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
			endIndex = a.len
		}
		go func(start int, end int) {
			defer wg.Done()
			for i := range a.data[start:end] {
				if !yield(&a.data[i+startIndex], workedId) {
					return
				}
			}
		}(startIndex, endIndex)
	}
	wg.Wait()
}
