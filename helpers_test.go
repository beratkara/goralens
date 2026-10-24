package goralens

import (
	"context"
	"testing"
	"time"
)

func newTestRecorder(t *testing.T, configure ...func(*Config)) (*Recorder, *MemoryStorage) {
	t.Helper()
	cfg := defaultConfig()
	cfg.FlushInterval = 10 * time.Millisecond
	for _, fn := range configure {
		fn(cfg)
	}
	storage := NewMemoryStorage(0)
	r := NewRecorder(cfg, storage)
	SetRecorder(r)
	t.Cleanup(func() {
		r.Close()
		SetRecorder(nil)
		hooksMu.Lock()
		hookSet = hooks{}
		hooksMu.Unlock()
	})
	return r, storage
}

func stored(t *testing.T, r *Recorder, s *MemoryStorage, typ EntryType) []*Entry {
	t.Helper()
	r.Flush(time.Second)
	entries, err := s.Get(context.Background(), Query{Type: typ, Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func one(t *testing.T, r *Recorder, s *MemoryStorage, typ EntryType) *Entry {
	t.Helper()
	entries := stored(t, r, s, typ)
	if len(entries) != 1 {
		t.Fatalf("expected 1 %s entry, got %d", typ, len(entries))
	}
	return entries[0]
}
