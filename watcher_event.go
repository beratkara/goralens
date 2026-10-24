package goralens

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/event"
)

func WrapEvent(e event.Instance) event.Instance {
	if e == nil {
		return nil
	}
	if _, ok := e.(*watchedEvents); ok {
		return e
	}
	return &watchedEvents{Instance: e}
}

type watchedEvents struct {
	event.Instance
}

func (w *watchedEvents) Job(ev event.Event, args []event.Arg) event.Task {
	return &watchedEventTask{Task: w.Instance.Job(ev, args), events: w.Instance, event: ev, args: args}
}

type watchedEventTask struct {
	event.Task
	events event.Instance
	event  event.Event
	args   []event.Arg
}

func (t *watchedEventTask) Dispatch() error {
	start := time.Now()
	err := t.Task.Dispatch()
	r := Instance()
	if r == nil || !r.config.watching(TypeEvent) {
		return err
	}
	payload := make([]map[string]any, len(t.args))
	for i, a := range t.args {
		payload[i] = map[string]any{"type": a.Type, "value": exportValue(a.Value)}
	}
	var listeners []map[string]any
	for _, l := range t.events.GetEvents()[t.event] {
		queued := l.Queue().Enable
		listeners = append(listeners, map[string]any{"name": l.Signature(), "queued": queued})
	}
	content := map[string]any{
		"name":      typeName(t.event),
		"payload":   payload,
		"listeners": listeners,
		"duration":  milliseconds(time.Since(start)),
	}
	if err != nil {
		content["error"] = err.Error()
	}
	r.Record(context.Background(), newEntry(TypeEvent, content))
	return err
}
