package goralens

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Recorder struct {
	config  *Config
	storage Storage

	recording atomic.Bool
	queue     chan []*Entry
	flushReq  chan chan struct{}
	stop      chan struct{}
	stopped   chan struct{}
	dropped   atomic.Int64

	monitorMu sync.RWMutex
	monitored map[string]bool

	errMu     sync.Mutex
	lastError time.Time
}

func NewRecorder(cfg *Config, storage Storage) *Recorder {
	if cfg == nil {
		cfg = defaultConfig()
	}
	r := &Recorder{
		config:    cfg,
		storage:   storage,
		queue:     make(chan []*Entry, max(cfg.QueueSize, 1)),
		flushReq:  make(chan chan struct{}),
		stop:      make(chan struct{}),
		stopped:   make(chan struct{}),
		monitored: map[string]bool{},
	}
	r.recording.Store(cfg.Enabled)
	r.loadMonitoredTags()
	go r.write()
	return r
}

func (r *Recorder) Config() *Config { return r.config }

func (r *Recorder) Storage() Storage { return r.storage }

func (r *Recorder) Recording() bool { return r.recording.Load() }

func (r *Recorder) SetRecording(on bool) { r.recording.Store(on) }

func (r *Recorder) Dropped() int64 { return r.dropped.Load() }

func (r *Recorder) Record(ctx context.Context, e *Entry) {
	if r == nil || e == nil || !r.recording.Load() || !r.config.watching(e.Type) {
		return
	}
	if skipped(ctx) {
		return
	}
	e.UUID = uuid.NewString()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	batch := batchFromContext(ctx)
	if batch == nil {
		state, ok := currentGoroutine()
		if ok && state.suppressed {
			return
		}
		batch = state.batch
	}
	if batch != nil {
		e.BatchID = batch.ID
		batch.add(e)
		return
	}
	e.BatchID = uuid.NewString()
	r.enqueue([]*Entry{e}, nil)
}

func (r *Recorder) StartBatch() (*Batch, func()) {
	b := newBatch()
	unbind := bindGoroutine(goroutineState{batch: b})
	return b, func() {
		unbind()
		entries, tags := b.drain()
		r.enqueue(entries, tags)
	}
}

func (r *Recorder) JoinOrStartBatch() (*Batch, func()) {
	if state, ok := currentGoroutine(); ok && state.batch != nil {
		return state.batch, func() {}
	}
	return r.StartBatch()
}

func (r *Recorder) Suppress() func() {
	return bindGoroutine(goroutineState{suppressed: true})
}

func (r *Recorder) enqueue(entries []*Entry, batchTags []string) {
	kept := entries[:0]
	for _, e := range entries {
		e.AddTags(batchTags...)
		if tagger := currentHooks().tag; tagger != nil {
			e.AddTags(tagger(e)...)
		}
		if r.keep(e) {
			kept = append(kept, e)
		}
	}
	if len(kept) == 0 {
		return
	}
	select {
	case r.queue <- kept:
	default:
		r.dropped.Add(int64(len(kept)))
	}
}

func (r *Recorder) keep(e *Entry) bool {
	if r.isMonitored(e.Tags) {
		return true
	}
	if filter := currentHooks().filter; filter != nil {
		return filter(e)
	}
	return true
}

func (r *Recorder) Flush(timeout time.Duration) {
	done := make(chan struct{})
	select {
	case r.flushReq <- done:
	case <-time.After(timeout):
		return
	case <-r.stopped:
		return
	}
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

func (r *Recorder) Close() {
	select {
	case <-r.stop:
		return
	default:
	}
	close(r.stop)
	<-r.stopped
}

func (r *Recorder) write() {
	defer close(r.stopped)
	defer r.Suppress()()
	ticker := time.NewTicker(r.config.FlushInterval)
	defer ticker.Stop()
	var pending []*Entry
	store := func() {
		if len(pending) > 0 {
			r.store(pending)
			pending = nil
		}
	}
	drain := func() {
		for {
			select {
			case entries := <-r.queue:
				pending = append(pending, entries...)
			default:
				return
			}
		}
	}
	for {
		select {
		case entries := <-r.queue:
			pending = append(pending, entries...)
			if len(pending) >= 500 {
				store()
			}
		case <-ticker.C:
			store()
		case done := <-r.flushReq:
			drain()
			store()
			close(done)
		case <-r.stop:
			drain()
			store()
			return
		}
	}
}

func (r *Recorder) store(entries []*Entry) {
	if r.storage == nil {
		return
	}

	seen := map[string]bool{}
	for i := len(entries) - 1; i >= 0; i-- {
		if hash := entries[i].FamilyHash; hash != "" {
			entries[i].DisplayOnIndex = !seen[hash]
			seen[hash] = true
		}
	}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), skipKey{}, true), 10*time.Second)
	defer cancel()
	if err := r.storage.Store(ctx, entries); err != nil {
		r.reportError(err)
	}
}

func (r *Recorder) reportError(err error) {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	if time.Since(r.lastError) < time.Minute {
		return
	}
	r.lastError = time.Now()
	fmt.Fprintf(os.Stderr, "goralens: storing entries failed: %v\n", err)
}

func (r *Recorder) loadMonitoredTags() {
	if r.storage == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), skipKey{}, true), 5*time.Second)
	defer cancel()
	tags, err := r.storage.MonitoredTags(ctx)
	if err != nil {
		return
	}
	r.setMonitored(tags)
}

func (r *Recorder) setMonitored(tags []string) {
	m := make(map[string]bool, len(tags))
	for _, t := range tags {
		m[t] = true
	}
	r.monitorMu.Lock()
	r.monitored = m
	r.monitorMu.Unlock()
}

func (r *Recorder) isMonitored(tags []string) bool {
	r.monitorMu.RLock()
	defer r.monitorMu.RUnlock()
	for _, t := range tags {
		if r.monitored[t] {
			return true
		}
	}
	return false
}
