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

type PagedArray[T any] struct {
	book             []ArrayPage[T]
	currentPageIndex int
	len              int
	parallelCount    uint8
}

func NewPagedArray[T any]() (a PagedArray[T]) {
	a.book = make([]ArrayPage[T], 2, book_size)
	a.parallelCount = uint8(runtime.NumCPU()) / 2

	return a
}

func (a *PagedArray[T]) Len() int {
	return a.len
}

func (a *PagedArray[T]) Get(index int) *T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	pageId, index := a.getPageIdAndIndex(index)
	page := &a.book[pageId]

	return &(page.data[index])
}

func (a *PagedArray[T]) GetValue(index int) T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	pageId, index := a.getPageIdAndIndex(index)
	page := &a.book[pageId]

	return (page.data[index])
}

func (a *PagedArray[T]) Set(index int, value T) *T {
	assert.True(index >= 0, "index out of range")
	assert.True(index < a.len, "index out of range")

	pageId, index := a.getPageIdAndIndex(index)
	page := &a.book[pageId]

	page.data[index] = value

	return &page.data[index]
}

func (a *PagedArray[T]) Append(value T) *T {
	if a.currentPageIndex >= len(a.book) {
		newBooks := make([]ArrayPage[T], len(a.book)*2)
		a.book = append(a.book, newBooks...)
	}

	page := &a.book[a.currentPageIndex]

	page.data[page.len] = value
	result := &page.data[page.len]
	page.len++
	if page.len == page_size {
		a.currentPageIndex++
	}
	a.len++
	return result
}

func (a *PagedArray[T]) SoftReduce() {
	assert.True(a.len > 0, "Len is already 0")

	page := &a.book[a.currentPageIndex]
	assert.True(page.len > 0, "Len is already 0")

	page.len--
	a.len--

	if page.len == 0 && a.currentPageIndex != 0 {
		a.currentPageIndex--
	}
}

func (a *PagedArray[T]) Reset() {
	for range a.currentPageIndex {
		page := &a.book[a.currentPageIndex]
		page.len = 0
	}

	a.currentPageIndex = 0
	a.len = 0
}

func (a *PagedArray[T]) Copy(fromIndex, toIndex int) {
	assert.True(fromIndex >= 0, "index out of range")
	assert.True(fromIndex < a.len, "index out of range")
	from := a.Get(fromIndex)

	assert.True(toIndex >= 0, "index out of range")
	assert.True(toIndex < a.len, "index out of range")
	to := a.Get(toIndex)

	*to = *from
}

func (a *PagedArray[T]) Swap(i, j int) (newI, NewJ *T) {
	assert.True(i >= 0, "index out of range")
	assert.True(i < a.len, "index out of range")
	x := a.Get(i)

	assert.True(j >= 0, "index out of range")
	assert.True(j < a.len, "index out of range")
	y := a.Get(j)

	*x, *y = *y, *x
	return x, y
}

func (a *PagedArray[T]) Last() *T {
	index := a.len - 1
	assert.True(index >= 0, "index out of range")

	return a.Get(index)
}

func (a *PagedArray[T]) Raw(result []T) {
	result = result[:0]
	for i := 0; i <= a.currentPageIndex; i++ {
		page := &a.book[i]
		result = append(result[:i*1024], append(result[i*1024:], page.data[:page.len]...)...)
	}
}

func (a *PagedArray[T]) getPageIdAndIndex(index int) (int, int) {
	pageId := index >> page_size_shift
	assert.True(pageId < len(a.book), "index out of range")

	index %= page_size
	assert.True(index < page_size, "index out of range")

	return pageId, index
}

func (a *PagedArray[T]) All(yield func(int, *T) bool) {
	var page *ArrayPage[T]
	var index_offset int

	book := a.book

	if a.len == 0 {
		return
	}

	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]
		index_offset = i << page_size_shift

		for j := page.len - 1; j >= 0; j-- {
			if !yield(index_offset+j, &page.data[j]) {
				return
			}
		}
	}
}

func (a *PagedArray[T]) AllParallel(yield func(int, *T) bool) {
	var page *ArrayPage[T]
	var data *[page_size]T
	var index_offset int

	book := a.book
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount

	runner := func(data *[page_size]T, offset int, startIndex int, wg *sync.WaitGroup) {
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
		index_offset = int(i) << page_size_shift

		if gorutineBudget > 0 {
			go runner(data, index_offset, page.len-1, wg)
			gorutineBudget--
			continue
		}

		runner(data, index_offset, page.len-1, wg)
	}

	wg.Wait()
}

func (a *PagedArray[T]) AllData(yield func(*T) bool) {
	var page *ArrayPage[T]

	book := a.book

	if a.len == 0 {
		return
	}

	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]

		for j := page.len - 1; j >= 0; j-- {
			if !yield(&page.data[j]) {
				return
			}
		}
	}
}

func (a *PagedArray[T]) AllDataValue(yield func(T) bool) {
	var page *ArrayPage[T]

	book := a.book

	if a.len == 0 {
		return
	}

	for i := a.currentPageIndex; i >= 0; i-- {
		page = &book[i]

		for j := page.len - 1; j >= 0; j-- {
			if !yield(page.data[j]) {
				return
			}
		}
	}
}

func (a *PagedArray[T]) AllDataValueParallel(yield func(T) bool) {
	var page *ArrayPage[T]
	var data *[page_size]T

	book := a.book
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount
	runner := func(data *[page_size]T, startIndex int, wg *sync.WaitGroup) {
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

func (a *PagedArray[T]) AllDataParallel(yield func(*T) bool) {
	var page *ArrayPage[T]
	var data *[page_size]T

	book := a.book
	wg := new(sync.WaitGroup)
	gorutineBudget := a.parallelCount
	runner := func(data *[page_size]T, startIndex int, wg *sync.WaitGroup) {
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

type SlicePage[T any] struct {
	len  int
	data []T
}

type ArrayPage[T any] struct {
	len  int
	data [page_size]T
}
