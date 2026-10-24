package goralens

import (
	"github.com/goravel/framework/contracts/log"
)

func LogDriver() log.Logger {
	return logDriver{}
}

type logDriver struct{}

func (logDriver) Handle(string) (log.Handler, error) {
	return logHandler{}, nil
}

type logHandler struct{}

func (logHandler) Enabled(level log.Level) bool {
	r := Instance()
	if r == nil || !r.Recording() {
		return false
	}
	min, err := log.ParseLevel(r.config.LogLevel)
	if err != nil {
		min = log.LevelDebug
	}
	if level < min && !(level >= log.LevelError && r.config.ExceptionsFromLogs) {
		return false
	}
	return r.config.watching(TypeLog) || r.config.watching(TypeException)
}

func (logHandler) Handle(entry log.Entry) error {
	r := Instance()
	if r == nil {
		return nil
	}
	ctx := entry.Context()

	context := map[string]any{}
	for k, v := range entry.With() {
		context[k] = exportValue(v)
	}
	for k, v := range entry.Data() {
		context[k] = exportValue(v)
	}
	content := map[string]any{"level": entry.Level().String(), "message": entry.Message(), "context": context}
	if code := entry.Code(); code != "" {
		content["code"] = code
	}
	if domain := entry.Domain(); domain != "" {
		content["domain"] = domain
	}
	if hint := entry.Hint(); hint != "" {
		content["hint"] = hint
	}
	if user := entry.User(); user != nil {
		content["user"] = exportValue(user)
	}
	if req := entry.Request(); len(req) > 0 {
		content["request"] = exportValue(req)
	}
	min, _ := log.ParseLevel(r.config.LogLevel)
	if r.config.watching(TypeLog) && entry.Level() >= min {
		e := newEntry(TypeLog, content)
		e.CreatedAt = entry.Time()
		e.AddTags(entry.Tags()...)
		r.Record(ctx, e)
	}
	if r.config.ExceptionsFromLogs && entry.Level() >= log.LevelError {
		extra := map[string]any{"class": "Log " + entry.Level().String(), "level": entry.Level().String(), "context": context}
		if trace := entry.Trace(); len(trace) > 0 {
			extra["log_trace"] = exportValue(trace)
		}
		r.recordException(ctx, entry.Message(), captureStack(1), extra)
	}
	return nil
}
