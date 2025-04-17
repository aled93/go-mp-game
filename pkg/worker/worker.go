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
	"github.com/negrel/assert"
	"runtime"
	"sync"
)

type WorkerId int

func NewWorker(ctx context.Context, id WorkerId) Worker {
	return Worker{
		id:             id,
		ctx:            ctx,
		groupTasksChan: make(chan groupTasks),
	}
}

type Worker struct {
	id             WorkerId
	ctx            context.Context
	groupTasksChan chan groupTasks
	wg             sync.WaitGroup
}

func (w *Worker) Start(poolWg *sync.WaitGroup) {
	poolWg.Add(1)
	go w.run(poolWg)
}

func (w *Worker) Stop() {
	close(w.groupTasksChan)
}

func (w *Worker) run(poolWg *sync.WaitGroup) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer poolWg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case job, ok := <-w.groupTasksChan:
			if !ok {
				return
			}
			w.processGroupTasks(job)
		}
	}
}

func (w *Worker) processGroupTasks(job groupTasks) {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-job.ctx.Done():
			return
		case task := <-job.taskChan:
			w.processTask(job, task)
		}
	}
}

func (w *Worker) processTask(job groupTasks, task AnyTask) {
	defer job.wg.Done()
	assert.NotNil(job.ctx)
	assert.NotNil(task)

	err := task.Run(job.ctx, w.id)
	if err != nil {
		job.errChan <- TaskError{Err: err, Id: w.id}
		return
	}
}

func (w *Worker) Id() WorkerId {
	return w.id
}
