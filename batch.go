package goralens

import (
	"bytes"
	"context"
	"runtime"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type Batch struct {
	ID string

	mu      sync.Mutex
	entries []*Entry
	tags    []string
}

func newBatch() *Batch {
	return &Batch{ID: uuid.NewString()}
}

func (b *Batch) add(e *Entry) {
	b.mu.Lock()
	b.entries = append(b.entries, e)
	b.mu.Unlock()
}

func (b *Batch) Tag(tags ...string) {
	b.mu.Lock()
	b.tags = append(b.tags, tags...)
	b.mu.Unlock()
}

func (b *Batch) drain() ([]*Entry, []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entries, tags := b.entries, b.tags
	b.entries, b.tags = nil, nil
	return entries, tags
}

type batchKey struct{}

type skipKey struct{}

func WithBatch(ctx context.Context, b *Batch) context.Context {
	return context.WithValue(ctx, batchKey{}, b)
}

func batchFromContext(ctx context.Context) *Batch {
	if ctx == nil {
		return nil
	}
	b, _ := ctx.Value(batchKey{}).(*Batch)
	return b
}

func skipped(ctx context.Context) bool {
	return ctx != nil && ctx.Value(skipKey{}) != nil
}

type goroutineState struct {
	batch      *Batch
	suppressed bool
}

var goroutines sync.Map

func bindGoroutine(state goroutineState) (unbind func()) {
	id := goroutineID()
	previous, had := goroutines.Load(id)
	goroutines.Store(id, state)
	return func() {
		if had {
			goroutines.Store(id, previous)
		} else {
			goroutines.Delete(id)
		}
	}
}

func currentGoroutine() (goroutineState, bool) {
	v, ok := goroutines.Load(goroutineID())
	if !ok {
		return goroutineState{}, false
	}
	return v.(goroutineState), true
}

func goroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	b := bytes.TrimPrefix(buf[:n], []byte("goroutine "))
	if i := bytes.IndexByte(b, ' '); i > 0 {
		b = b[:i]
	}
	id, _ := strconv.ParseUint(string(b), 10, 64)
	return id
}
