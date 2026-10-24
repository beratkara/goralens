package goralens

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStorageQueries(t *testing.T) {
	s := NewMemoryStorage(0)
	ctx := context.Background()
	old := &Entry{UUID: "a", BatchID: "b1", Type: TypeLog, Tags: []string{"x"}, CreatedAt: time.Now().Add(-48 * time.Hour), DisplayOnIndex: true}
	fresh := &Entry{UUID: "b", BatchID: "b1", Type: TypeQuery, CreatedAt: time.Now(), DisplayOnIndex: true}
	if err := s.Store(ctx, []*Entry{old, fresh}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(ctx, Query{Tag: "x"}); len(got) != 1 || got[0].UUID != "a" {
		t.Fatalf("tag query: %+v", got)
	}
	if got, _ := s.Get(ctx, Query{BatchID: "b1"}); len(got) != 2 || got[0].UUID != "b" {
		t.Fatalf("batch query should return newest first: %+v", got)
	}
	if got, _ := s.Get(ctx, Query{Before: 2}); len(got) != 1 || got[0].UUID != "a" {
		t.Fatalf("paging: %+v", got)
	}
	if e, _ := s.Find(ctx, "b"); e == nil || e.Type != TypeQuery {
		t.Fatalf("find: %+v", e)
	}
	if n, _ := s.Prune(ctx, time.Now().Add(-24*time.Hour)); n != 1 {
		t.Fatalf("pruned %d", n)
	}
	if err := s.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	if counts, _ := s.Counts(ctx); len(counts) != 0 {
		t.Fatalf("not cleared: %v", counts)
	}
}

func TestExceptionsAreGroupedByFamily(t *testing.T) {
	r, s := newTestRecorder(t)
	for i := 0; i < 3; i++ {
		r.RecordException(context.Background(), errTest("boom"), nil)
	}
	r.Flush(time.Second)
	ctx := context.Background()
	listed, _ := s.Get(ctx, Query{Type: TypeException, IndexOnly: true})
	if len(listed) != 1 {
		t.Fatalf("index shows %d occurrences, want 1", len(listed))
	}
	if n, _ := s.Occurrences(ctx, listed[0].FamilyHash); n != 3 {
		t.Fatalf("occurrences = %d", n)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }
