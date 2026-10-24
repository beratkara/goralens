package goralens

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/queue"
)

func WrapQueue(q queue.Queue) queue.Queue {
	if q == nil {
		return nil
	}
	if _, ok := q.(*watchedQueue); ok {
		return q
	}
	return &watchedQueue{Queue: q}
}

type watchedQueue struct {
	queue.Queue
}

func (w *watchedQueue) Register(jobs []queue.Job) {
	wrapped := make([]queue.Job, len(jobs))
	for i, j := range jobs {
		wrapped[i] = watchJob(j)
	}
	w.Queue.Register(wrapped)
}

func (w *watchedQueue) Job(job queue.Job, args ...[]queue.Arg) queue.PendingJob {
	var jobArgs []queue.Arg
	if len(args) > 0 {
		jobArgs = args[0]
	}
	return &watchedPendingJob{
		PendingJob: w.Queue.Job(watchJob(job), args...),
		jobs:       []queue.ChainJob{{Job: job, Args: jobArgs}},
	}
}

func (w *watchedQueue) Chain(jobs []queue.ChainJob) queue.PendingJob {
	wrapped := make([]queue.ChainJob, len(jobs))
	for i, j := range jobs {
		wrapped[i] = j
		wrapped[i].Job = watchJob(j.Job)
	}
	return &watchedPendingJob{PendingJob: w.Queue.Chain(wrapped), jobs: jobs}
}

type watchedPendingJob struct {
	queue.PendingJob
	jobs       []queue.ChainJob
	connection string
	queueName  string
	delay      time.Time
}

func (p *watchedPendingJob) Delay(t time.Time) queue.PendingJob {
	p.PendingJob = p.PendingJob.Delay(t)
	p.delay = t
	return p
}

func (p *watchedPendingJob) OnConnection(connection string) queue.PendingJob {
	p.PendingJob = p.PendingJob.OnConnection(connection)
	p.connection = connection
	return p
}

func (p *watchedPendingJob) OnQueue(name string) queue.PendingJob {
	p.PendingJob = p.PendingJob.OnQueue(name)
	p.queueName = name
	return p
}

func (p *watchedPendingJob) Dispatch() error {
	err := p.PendingJob.Dispatch()
	p.record(err, false)
	return err
}

func (p *watchedPendingJob) DispatchSync() error {
	err := p.PendingJob.DispatchSync()
	p.record(err, true)
	return err
}

func (p *watchedPendingJob) record(err error, sync bool) {
	r := Instance()
	if r == nil || !r.config.watching(TypeJob) {
		return
	}
	for i, j := range p.jobs {
		content := map[string]any{
			"status":     "queued",
			"name":       j.Job.Signature(),
			"connection": p.connection,
			"queue":      p.queueName,
			"data":       jobArgs(j.Args),
			"sync":       sync,
		}
		if len(p.jobs) > 1 {
			content["chain_position"] = i + 1
		}
		if !p.delay.IsZero() {
			content["delay_until"] = p.delay
		}
		if err != nil {
			content["status"] = "failed"
			content["error"] = err.Error()
		}
		e := newEntry(TypeJob, content)
		e.AddTags(content["status"].(string))
		r.Record(context.Background(), e)
	}
}

func jobArgs(args []queue.Arg) []map[string]any {
	out := make([]map[string]any, len(args))
	for i, a := range args {
		out[i] = map[string]any{"type": a.Type, "value": exportValue(a.Value)}
	}
	return out
}

func watchJob(j queue.Job) queue.Job {
	if j == nil {
		return nil
	}
	switch j.(type) {
	case *watchedJob, *watchedRetryJob:
		return j
	}
	if retry, ok := j.(queue.JobWithShouldRetry); ok {
		return &watchedRetryJob{watchedJob: watchedJob{Job: j}, retry: retry}
	}
	return &watchedJob{Job: j}
}

type watchedJob struct {
	queue.Job
}

func (j *watchedJob) Handle(args ...any) (err error) {
	r := Instance()
	if r == nil || !r.Recording() {
		return j.Job.Handle(args...)
	}
	start := time.Now()
	_, end := r.JoinOrStartBatch()
	defer func() {
		recovered := recover()
		if recovered != nil {
			r.recordException(context.Background(), recovered, panicTrace(), map[string]any{"panic": true, "job": j.Job.Signature()})
			err = fmt.Errorf("job %s panicked: %v", j.Job.Signature(), recovered)
		}
		if r.config.watching(TypeJob) {
			data := make([]any, len(args))
			for i, a := range args {
				data[i] = exportValue(a)
			}
			content := map[string]any{
				"status":   "processed",
				"name":     j.Job.Signature(),
				"data":     data,
				"duration": milliseconds(time.Since(start)),
			}
			if err != nil {
				content["status"] = "failed"
				content["error"] = err.Error()
			}
			e := newEntry(TypeJob, content)
			e.CreatedAt = start
			e.AddTags(content["status"].(string))
			r.Record(context.Background(), e)
		}
		end()
		if recovered != nil {
			panic(recovered)
		}
	}()
	return j.Job.Handle(args...)
}

type watchedRetryJob struct {
	watchedJob
	retry queue.JobWithShouldRetry
}

func (j *watchedRetryJob) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	return j.retry.ShouldRetry(err, attempt)
}
