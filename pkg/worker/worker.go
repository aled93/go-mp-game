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

type WorkerId int

func NewWorker(ctx context.Context, id WorkerId) Worker {
	return Worker{
		id:      id,
		ctx:     ctx,
		jobChan: make(chan poolJob),
	}
}

type Worker struct {
	id      WorkerId
	ctx     context.Context
	jobChan chan poolJob
	wg      sync.WaitGroup
}

func (w *Worker) Start(poolWg *sync.WaitGroup) {
	poolWg.Add(1)
	go w.run(poolWg)
}

func (w *Worker) Stop() {
	close(w.jobChan)
}

func (w *Worker) run(poolWg *sync.WaitGroup) {
	defer poolWg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case job, ok := <-w.jobChan:
			if !ok {
				return
			}
			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job poolJob) {
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

func (w *Worker) processTask(job poolJob, task AnyTask) {
	defer job.wg.Done()
	err := task.Run(job.ctx, w.id)
	if err != nil {
		job.errChan <- TaskError{Err: err, Id: w.id}
		return
	}
}

func (w *Worker) Id() WorkerId {
	return w.id
}
