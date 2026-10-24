package goralens

import (
	"context"
	"time"
)

type Query struct {
	Type       EntryType
	Tag        string
	BatchID    string
	FamilyHash string

	Before int64
	Limit  int

	IndexOnly bool
}

type Storage interface {
	Store(ctx context.Context, entries []*Entry) error
	Get(ctx context.Context, q Query) ([]*Entry, error)
	Find(ctx context.Context, uuid string) (*Entry, error)

	Occurrences(ctx context.Context, familyHash string) (int64, error)

	Counts(ctx context.Context) (map[EntryType]int64, error)

	Prune(ctx context.Context, before time.Time) (int64, error)
	Clear(ctx context.Context) error

	MonitoredTags(ctx context.Context) ([]string, error)
	Monitor(ctx context.Context, tags []string) error
	Unmonitor(ctx context.Context, tags []string) error
}

func (q Query) limit() int {
	if q.Limit <= 0 || q.Limit > 200 {
		return 50
	}
	return q.Limit
}
