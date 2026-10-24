package goralens

import (
	"context"
	"time"
)

func RecordRedis(ctx context.Context, connection, command string, duration time.Duration, err error) {
	r := Instance()
	if r == nil || !r.config.watching(TypeRedis) {
		return
	}
	content := map[string]any{
		"connection": connection,
		"command":    command,
		"time":       milliseconds(duration),
	}
	if err != nil {
		content["error"] = err.Error()
	}
	r.Record(ctx, newEntry(TypeRedis, content))
}
