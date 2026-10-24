package goralens

import (
	"context"
	"encoding/json"
	"fmt"
)

func Dump(values ...any) {
	DumpContext(context.Background(), values...)
}

func DumpContext(ctx context.Context, values ...any) {
	r := Instance()
	if r == nil || !r.config.watching(TypeDump) {
		return
	}
	for _, v := range values {
		content := map[string]any{"dump": dumpString(v), "type": fmt.Sprintf("%T", v)}
		if f, ok := origin(captureStack(1)); ok {
			content["file"], content["line"] = f.File, f.Line
		}
		r.Record(ctx, newEntry(TypeDump, content))
	}
}

func dumpString(v any) string {
	if b, err := json.MarshalIndent(exportValue(v), "", "  "); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%#v", v)
}
