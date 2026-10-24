package goralens

import (
	"context"

	"github.com/goravel/framework/contracts/auth/access"
)

func WatchGate(gate access.Gate) {
	if gate == nil {
		return
	}
	gate.After(func(ctx context.Context, ability string, arguments map[string]any, result access.Response) access.Response {
		r := Instance()
		if r == nil || !r.config.watching(TypeGate) || matchesAny(ability, r.config.IgnoreAbilities) {
			return nil
		}
		content := map[string]any{"ability": ability, "result": "denied", "arguments": exportValue(arguments)}
		if result != nil {
			if result.Allowed() {
				content["result"] = "allowed"
			}
			if msg := result.Message(); msg != "" {
				content["message"] = msg
			}
		}
		if f, ok := origin(captureStack(1)); ok {
			content["file"], content["line"] = f.File, f.Line
		}
		r.Record(ctx, newEntry(TypeGate, content))

		return nil
	})
}
