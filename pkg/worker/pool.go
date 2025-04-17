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

package worker

import (
	"context"
	"sync"
)

func NewPool(n int) Pool {
	return Pool{
		workers:  make([]Worker, n),
		taskChan: make(chan AnyTask),
		errChan:  make(chan TaskError),
	}
}

type Pool struct {
	workers   []Worker
	wg        sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc

	// Cache
	taskChan    chan AnyTask
	errChan     chan TaskError
	groupTaskWg sync.WaitGroup
}

func (p *Pool) Start() {
	p.ctx, p.ctxCancel = context.WithCancel(context.Background())
	for i := range p.workers {
		p.workers[i] = NewWorker(p.ctx, WorkerId(i))
		p.workers[i].Start(&p.wg)
	}
}

func (p *Pool) AddWorker() {
	p.workers = append(p.workers, NewWorker(p.ctx, WorkerId(len(p.workers))))
	p.workers[len(p.workers)-1].Start(&p.wg)
}

func (p *Pool) RemoveWorker() {
	p.workers[len(p.workers)-1].Stop()
	p.workers = p.workers[:len(p.workers)-1]
}

func (p *Pool) Stop() {
	p.ctxCancel()
	p.wg.Wait()
}

type groupTasks struct {
	taskChan <-chan AnyTask
	errChan  chan<- TaskError
	ctx      context.Context
	wg       *sync.WaitGroup
}

func (p *Pool) ProcessGroupTasks(tasks []AnyTask) error {
	var ctx, cancel = context.WithCancel(p.ctx)
	defer cancel()

	var job = groupTasks{
		taskChan: p.taskChan,
		errChan:  p.errChan,
		ctx:      ctx,
		wg:       &p.groupTaskWg,
	}

	for i := range p.workers {
		p.workers[i].groupTasksChan <- job
	}

	p.groupTaskWg.Add(len(tasks))
	for i := range tasks {
		select {
		case err := <-p.errChan:
			return err
		default:
			p.taskChan <- tasks[i]
		}
	}
	p.groupTaskWg.Wait()

	return nil
}

func (p *Pool) NumWorkers() int {
	return len(p.workers)
}
