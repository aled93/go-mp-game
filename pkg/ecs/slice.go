/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package ecs

import (
	"runtime"
	"sync"

	"github.com/negrel/assert"
)

type Slice[T any] struct {
	data          []T
	len           int
	parallelCount uint8
}

func NewSlice[T any](size int) (a Slice[T]) {
	a.data = make([]T, 0, size)
	a.parallelCount = uint8(runtime.NumCPU()) - 2

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
	a.data = append(a.data[:a.len-1], values...)
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

func (a *Slice[T]) Raw(result []T) []T {
	result = result[:0]
	copy(result, a.data)
	return result
}

func (a *Slice[T]) getPageIdAndIndex(index int) (int, int) {
	pageId := index >> pageSizeShift
	assert.True(pageId < len(a.data), "index out of range")

	index %= pageSize
	assert.True(index < pageSize, "index out of range")

	return pageId, index
}

func (a *Slice[T]) All(yield func(int, *T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(j, &a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) AllData(yield func(*T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(&a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) AllDataValue(yield func(T) bool) {
	for j := a.len - 1; j >= 0; j-- {
		if !yield(a.data[j]) {
			return
		}
	}
}

func (a *Slice[T]) AllParallel(yield func(int, *T) bool) {
	var page *ArrayPage[T]
	var data *[pageSize]T
	var index_offset int

	book := a.data
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount

	runner := func(data *[pageSize]T, offset int, startIndex int, wg *sync.WaitGroup) {
		defer wg.Done()
		for j := startIndex; j >= 0; j-- {
			if !yield(offset+j, &(data[j])) {
				return
			}
		}
	}

	if a.len == 0 {
		return
	}

	wg.Add(int(a.currentPageIndex) + 1)
	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]
		data = &page.data
		index_offset = int(i) << pageSizeShift

		if gorutineBudget > 0 {
			go runner(data, index_offset, page.len-1, wg)
			gorutineBudget--
			continue
		}

		runner(data, index_offset, page.len-1, wg)
	}

	wg.Wait()
}

func (a *Slice[T]) AllDataValueParallel(yield func(T) bool) {
	var page *ArrayPage[T]
	var data *[pageSize]T

	book := a.data
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount
	runner := func(data *[pageSize]T, startIndex int, wg *sync.WaitGroup) {
		defer wg.Done()
		for j := startIndex; j >= 0; j-- {
			if !yield((data[j])) {
				return
			}
		}
	}

	if a.len == 0 {
		return
	}

	wg.Add(int(a.currentPageIndex) + 1)
	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]
		data = &page.data

		if gorutineBudget > 0 {
			go runner(data, page.len-1, wg)
			gorutineBudget--
			continue
		}

		runner(data, page.len-1, wg)
	}
	wg.Wait()
}

func (a *Slice[T]) AllDataParallel(yield func(*T) bool) {
	var page *ArrayPage[T]
	var data *[pageSize]T

	book := a.data
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount
	runner := func(data *[pageSize]T, startIndex int, wg *sync.WaitGroup) {
		defer wg.Done()
		for j := startIndex; j >= 0; j-- {
			if !yield(&(data[j])) {
				return
			}
		}
	}

	if a.len == 0 {
		return
	}

	wg.Add(int(a.currentPageIndex) + 1)
	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]
		data = &page.data

		if gorutineBudget > 0 {
			go runner(data, page.len-1, wg)
			gorutineBudget--
			continue
		}

		runner(data, page.len-1, wg)
	}
	wg.Wait()
}
