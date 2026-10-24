package goralens

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEntriesOfABatchShareItsID(t *testing.T) {
	r, s := newTestRecorder(t)
	_, end := r.StartBatch()
	Dump("first")
	r.Record(context.Background(), newEntry(TypeLog, map[string]any{"message": "second"}))

	r.Flush(time.Second)
	if n, _ := s.Counts(context.Background()); len(n) != 0 {
		t.Fatalf("entries written before the batch ended: %v", n)
	}
	end()

	dump, log := one(t, r, s, TypeDump), one(t, r, s, TypeLog)
	if dump.BatchID == "" || dump.BatchID != log.BatchID {
		t.Fatalf("batch ids differ: %q %q", dump.BatchID, log.BatchID)
	}
}

func TestBatchFromContextWinsOverGoroutine(t *testing.T) {
	r, s := newTestRecorder(t)
	b := newBatch()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		DumpContext(WithBatch(context.Background(), b), "from goroutine")
	}()
	wg.Wait()
	entries, _ := b.drain()
	if len(entries) != 1 || entries[0].BatchID != b.ID {
		t.Fatalf("entry not added to the context batch: %+v", entries)
	}
	if got := stored(t, r, s, TypeDump); len(got) != 0 {
		t.Fatalf("batched entry written early")
	}
}

func TestEntriesWithoutBatchAreWrittenAtOnce(t *testing.T) {
	r, s := newTestRecorder(t)
	Dump(map[string]any{"a": 1})
	e := one(t, r, s, TypeDump)
	if e.BatchID == "" || e.UUID == "" {
		t.Fatalf("missing ids: %+v", e)
	}
}

func TestSuppressedGoroutineRecordsNothing(t *testing.T) {
	r, s := newTestRecorder(t)
	resume := r.Suppress()
	Dump("hidden")
	resume()
	Dump("visible")
	entries := stored(t, r, s, TypeDump)
	if len(entries) != 1 || entries[0].Content["dump"] != `"visible"` {
		t.Fatalf("unexpected dumps: %+v", entries)
	}
}

func TestSkipContextIsNotRecorded(t *testing.T) {
	r, s := newTestRecorder(t)
	DumpContext(context.WithValue(context.Background(), skipKey{}, true), "skipped")
	if got := stored(t, r, s, TypeDump); len(got) != 0 {
		t.Fatalf("skipped context recorded")
	}
}

func TestPausedAndDisabledWatchers(t *testing.T) {
	r, s := newTestRecorder(t, func(c *Config) { c.Watchers[TypeLog] = false })
	r.Record(context.Background(), newEntry(TypeLog, map[string]any{}))
	r.SetRecording(false)
	Dump("paused")
	r.SetRecording(true)
	if got := stored(t, r, s, ""); len(got) != 0 {
		t.Fatalf("recorded %d entries", len(got))
	}
}

func TestFilterTagAndMonitoring(t *testing.T) {
	r, s := newTestRecorder(t)
	Tag(func(e *Entry) []string { return []string{"custom"} })
	Filter(func(e *Entry) bool { return e.Content["keep"] == true })

	r.Record(context.Background(), newEntry(TypeLog, map[string]any{"keep": true}))
	r.Record(context.Background(), newEntry(TypeLog, map[string]any{"keep": false}))
	entries := stored(t, r, s, TypeLog)
	if len(entries) != 1 || !entries[0].HasTag("custom") {
		t.Fatalf("filter or tag not applied: %+v", entries)
	}

	if err := s.Monitor(context.Background(), []string{"custom"}); err != nil {
		t.Fatal(err)
	}
	r.loadMonitoredTags()
	r.Record(context.Background(), newEntry(TypeLog, map[string]any{"keep": false}))
	if got := stored(t, r, s, TypeLog); len(got) != 2 {
		t.Fatalf("monitored entry dropped, have %d", len(got))
	}
}

func TestBatchTagsReachEveryEntry(t *testing.T) {
	r, s := newTestRecorder(t)
	b, end := r.StartBatch()
	Dump(1)
	Dump(2)
	b.Tag("Auth:7")
	end()
	for _, e := range stored(t, r, s, TypeDump) {
		if !e.HasTag("Auth:7") {
			t.Fatalf("entry without batch tag: %+v", e.Tags)
		}
	}
}

func TestFullQueueDropsInsteadOfBlocking(t *testing.T) {
	cfg := defaultConfig()
	cfg.QueueSize = 1
	cfg.FlushInterval = time.Hour
	r := &Recorder{config: cfg, queue: make(chan []*Entry, 1), monitored: map[string]bool{}}
	r.recording.Store(true)

	r.enqueue([]*Entry{newEntry(TypeDump, nil)}, nil)
	r.enqueue([]*Entry{newEntry(TypeDump, nil)}, nil)
	if r.Dropped() != 1 {
		t.Fatalf("dropped = %d", r.Dropped())
	}
}

func TestGoroutineID(t *testing.T) {
	ids := make(chan uint64, 2)
	go func() { ids <- goroutineID() }()
	go func() { ids <- goroutineID() }()
	a, b := <-ids, <-ids
	if a == 0 || b == 0 || a == b {
		t.Fatalf("goroutine ids %d %d", a, b)
	}
}
