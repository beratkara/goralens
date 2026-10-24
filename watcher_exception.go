package goralens

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func internalFrame(fn, file string) bool {
	return strings.HasPrefix(fn, "runtime.") ||
		strings.Contains(file, "/goravel/framework") ||
		strings.Contains(file, "/goravel/gin") ||
		strings.Contains(file, "/goravel/fiber") ||
		strings.Contains(file, "/gin-gonic/") ||
		strings.Contains(file, "/gin-contrib/") ||
		strings.Contains(file, "gorm.io/") ||
		strings.Contains(file, "/sirupsen/logrus") ||
		strings.Contains(fn, "beratkara/goralens.") ||
		strings.Contains(fn, "beratkara/goralens/redishook.")
}

func captureStack(skip int) []Frame {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(skip+2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var out []Frame
	for {
		f, more := frames.Next()
		out = append(out, Frame{Function: f.Function, File: f.File, Line: f.Line})
		if !more {
			break
		}
	}
	return out
}

func origin(frames []Frame) (Frame, bool) {
	for _, f := range frames {
		if !internalFrame(f.Function, f.File) && !dependencyFrame(f.File) {
			return f, true
		}
	}
	return Frame{}, false
}

func dependencyFrame(file string) bool {
	return strings.Contains(file, "/pkg/mod/") || strings.Contains(file, "/vendor/") ||
		strings.HasPrefix(file, runtime.GOROOT())
}

func (r *Recorder) RecordException(ctx context.Context, err any, extra map[string]any) {
	r.recordException(ctx, err, captureStack(1), extra)
}

func (r *Recorder) recordException(ctx context.Context, err any, trace []Frame, extra map[string]any) {
	if r == nil || !r.config.watching(TypeException) {
		return
	}
	message := fmt.Sprint(err)
	class := typeName(err)
	if e, ok := err.(error); ok {
		message = e.Error()

		for inner := errors.Unwrap(e); inner != nil; inner = errors.Unwrap(inner) {
			class = typeName(inner)
		}
	}
	content := map[string]any{"class": class, "message": message, "trace": trace}
	if f, ok := origin(trace); ok {
		content["file"], content["line"] = f.File, f.Line
	}
	for k, v := range extra {
		content[k] = v
	}
	class = fmt.Sprint(content["class"])
	e := newEntry(TypeException, content)
	e.FamilyHash = exceptionFamily(class, content)
	e.AddTags(class)
	r.Record(ctx, e)
}

func exceptionFamily(class string, content map[string]any) string {
	key := class + "|" + fmt.Sprint(content["file"]) + ":" + fmt.Sprint(content["line"])
	if content["file"] == nil {
		key = class + "|" + fmt.Sprint(content["message"])
	}
	sum := sha1.Sum([]byte(key))
	return hex.EncodeToString(sum[:])
}

func panicTrace() []Frame {
	frames := captureStack(2)
	for i, f := range frames {
		if f.Function == "runtime.gopanic" || strings.HasSuffix(f.Function, "runtime.gopanic") {
			return frames[i+1:]
		}
	}
	return frames
}
