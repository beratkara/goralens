package goralens

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
)

func WrapSchedule(s schedule.Schedule, artisan console.Artisan) schedule.Schedule {
	if s == nil {
		return nil
	}
	if _, ok := s.(*watchedSchedule); ok {
		return s
	}
	return &watchedSchedule{Schedule: s, artisan: artisan}
}

type watchedSchedule struct {
	schedule.Schedule
	artisan console.Artisan
}

func (w *watchedSchedule) Register(events []schedule.Event) {
	wrapped := make([]schedule.Event, len(events))
	for i, e := range events {
		wrapped[i] = &watchedScheduleEvent{Event: e, artisan: w.artisan}
	}
	w.Schedule.Register(wrapped)
}

type watchedScheduleEvent struct {
	schedule.Event
	artisan console.Artisan
}

func (e *watchedScheduleEvent) GetCommand() string {
	if e.artisan == nil {
		return e.Event.GetCommand()
	}

	return ""
}

func (e *watchedScheduleEvent) GetCallback() func() {
	command := e.Event.GetCommand()
	callback := e.Event.GetCallback()
	return func() {
		r := Instance()
		var commandErr error
		run := callback
		if command != "" && e.artisan != nil {
			run = func() {
				if commandErr = e.artisan.Call(command); commandErr != nil {

					fmt.Fprintf(os.Stderr, "run %s command error: %v\n", command, commandErr)
				}
			}
		}
		if r == nil || !r.Recording() {
			if run != nil {
				run()
			}
			return
		}
		start := time.Now()
		_, end := r.JoinOrStartBatch()
		defer func() {
			recovered := recover()
			if r.config.watching(TypeSchedule) {
				name := e.Event.GetName()
				if name == "" {
					name = command
				}
				if name == "" {
					name = "Closure"
				}
				content := map[string]any{
					"name":       name,
					"command":    command,
					"expression": e.Event.GetCron(),
					"duration":   milliseconds(time.Since(start)),
					"status":     "processed",
				}
				if recovered != nil {
					content["status"] = "failed"
					content["error"] = fmt.Sprint(recovered)
				} else if commandErr != nil {
					content["status"] = "failed"
					content["error"] = commandErr.Error()
				}
				entry := newEntry(TypeSchedule, content)
				entry.CreatedAt = start
				r.Record(context.Background(), entry)
			}
			if recovered != nil {
				r.recordException(context.Background(), recovered, panicTrace(), map[string]any{"panic": true})
			}
			end()
			if recovered != nil {
				panic(recovered)
			}
		}()
		if run != nil {
			run()
		}
	}
}
