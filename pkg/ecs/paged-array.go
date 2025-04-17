/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package ecs

import (
	"context"
	"errors"
	"github.com/labstack/gommon/log"
	"github.com/negrel/assert"
	"gomp/pkg/worker"
	"sync"
)

type PagedArray[T any] struct {
	book             []ArrayPage[T]
	currentPageIndex int
	len              int
	wg               sync.WaitGroup
	edvpTasks        []EdvpTask[T]
	tasks            []worker.AnyTask
}

func NewPagedArray[T any]() (a PagedArray[T]) {
	a.book = make([]ArrayPage[T], 2, initialBookSize)
	a.edvpTasks = make([]EdvpTask[T], initialBookSize)
	a.tasks = make([]worker.AnyTask, initialBookSize)

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

func (a *PagedArray[T]) Append(values ...T) *T {
	var result *T
	for i := range values {
		value := values[i]
		if a.currentPageIndex >= len(a.book) {
			newBooks := make([]ArrayPage[T], len(a.book)*2)
			a.book = append(a.book, newBooks...)
			newEdvpTasks := make([]EdvpTask[T], len(a.edvpTasks)*2)
			a.edvpTasks = append(a.edvpTasks, newEdvpTasks...)
			newTasks := make([]worker.AnyTask, len(a.tasks)*2)
			a.tasks = append(a.tasks, newTasks...)
		}

		page := &a.book[a.currentPageIndex]

		if page.len == pageSize {
			a.currentPageIndex++
			if a.currentPageIndex >= len(a.book) {
				newBooks := make([]ArrayPage[T], len(a.book)*2)
				a.book = append(a.book, newBooks...)
				newEdvpTasks := make([]EdvpTask[T], len(a.edvpTasks)*2)
				a.edvpTasks = append(a.edvpTasks, newEdvpTasks...)
				newTasks := make([]worker.AnyTask, len(a.tasks)*2)
				a.tasks = append(a.tasks, newTasks...)
			}
			page = &a.book[a.currentPageIndex]
		}
		page.data[page.len] = value
		result = &page.data[page.len]
		page.len++
		a.len++
	}

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
	for i := 0; i <= a.currentPageIndex; i++ {
		page := &a.book[i]
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

func (a *PagedArray[T]) Raw(result []T) []T {
	result = result[:0]
	for i := 0; i <= a.currentPageIndex; i++ {
		page := &a.book[i]
		result = append(result[:i*1024], append(result[i*1024:], page.data[:page.len]...)...)
	}

	return result
}

func (a *PagedArray[T]) getPageIdAndIndex(index int) (int, int) {
	return index >> pageSizeShift, index % pageSize
}

func (a *PagedArray[T]) Each() func(yield func(int, *T) bool) {
	return func(yield func(int, *T) bool) {
		var page *ArrayPage[T]
		var index_offset int

		book := a.book

		if a.len == 0 {
			return
		}

		for i := a.currentPageIndex; i >= 0; i-- {
			page = &book[i]
			index_offset = i << pageSizeShift

			for j := page.len - 1; j >= 0; j-- {
				if !yield(index_offset+j, &page.data[j]) {
					return
				}
			}
		}
	}
}

func (a *PagedArray[T]) EachParallel(numWorkers int) func(yield func(int, *T, int) bool) {
	return func(yield func(int, *T, int) bool) {
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
				r := end - start
				for i := range r {
					if !yield(i, a.Get(i+startIndex), workedId) {
						return
					}
				}
			}(startIndex, endIndex)
		}
		wg.Wait()
	}
}

func (a *PagedArray[T]) EachData() func(yield func(*T) bool) {
	return func(yield func(*T) bool) {
		var page *ArrayPage[T]
		var book = a.book

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
}

func (a *PagedArray[T]) EachDataValue() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		var page *ArrayPage[T]
		var book = a.book

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
}

type EdvpTask[T any] struct {
	f    func(T, worker.WorkerId) bool
	page *ArrayPage[T]
}

func (t *EdvpTask[T]) Run(ctx context.Context, workerId worker.WorkerId) error {
	for i := 0; i < t.page.len; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			if shouldContinue := t.f(t.page.data[i], workerId); !shouldContinue {
				return errors.New("loop canceled")
			}
		}
	}
	return nil
}

func (a *PagedArray[T]) EachDataValueParallel(pool *worker.Pool) func(yield func(T, worker.WorkerId) bool) {
	return func(yield func(T, worker.WorkerId) bool) {
		assert.NotNil(pool)
		for i := a.currentPageIndex; i >= 0; i-- {
			j := a.currentPageIndex - i
			a.edvpTasks[j] = EdvpTask[T]{
				f:    yield,
				page: &a.book[i],
			}
			a.tasks[j] = &a.edvpTasks[j]
		}
		tasks := a.tasks[:a.currentPageIndex+1]
		err := pool.ProcessGroupTasks(tasks)
		if err != nil {
			log.Error(err)
		}
	}
}

func (a *PagedArray[T]) EachDataParallel(numWorkers int) func(yield func(*T, int) bool) {
	return func(yield func(*T, int) bool) {
		assert.True(numWorkers > 0)
		var chunkSize = a.len / numWorkers

		a.wg.Add(numWorkers)
		for workedId := 0; workedId < numWorkers; workedId++ {
			startIndex := workedId * chunkSize
			endIndex := startIndex + chunkSize - 1
			if workedId == numWorkers-1 { // have to set endIndex to entities length, if last worker
				endIndex = a.len
			}
			go a.edpTask(yield, startIndex, endIndex, workedId, &a.wg)
		}
		a.wg.Wait()
	}
}

func (a *PagedArray[T]) edpTask(y func(*T, int) bool, s int, e int, id int, w *sync.WaitGroup) {
	defer w.Done()
	r := e - s
	for i := range r {
		if !y(a.Get(i+s), id) {
			return
		}
	}
}

type SlicePage[T any] struct {
	len  int
	data []T
}

type ArrayPage[T any] struct {
	len  int
	data [pageSize]T
}
