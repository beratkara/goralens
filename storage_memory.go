package goralens

import (
	"context"
	"slices"
	"sync"
	"time"
)

type MemoryStorage struct {
	Limit int

	mu        sync.RWMutex
	entries   []*Entry
	sequence  int64
	monitored []string
}

func NewMemoryStorage(limit int) *MemoryStorage {
	if limit <= 0 {
		limit = 10000
	}
	return &MemoryStorage{Limit: limit}
}

func (m *MemoryStorage) Store(_ context.Context, entries []*Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range entries {
		if e.FamilyHash != "" {
			for _, old := range m.entries {
				if old.FamilyHash == e.FamilyHash {
					old.DisplayOnIndex = false
				}
			}
		}
		m.sequence++
		copied := *e
		copied.Sequence = m.sequence
		copied.Tags = slices.Clone(e.Tags)
		m.entries = append(m.entries, &copied)
	}
	if over := len(m.entries) - m.Limit; over > 0 {
		m.entries = slices.Delete(m.entries, 0, over)
	}
	return nil
}

func (m *MemoryStorage) Get(_ context.Context, q Query) ([]*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Entry
	for i := len(m.entries) - 1; i >= 0 && len(out) < q.limit(); i-- {
		e := m.entries[i]
		switch {
		case q.Type != "" && e.Type != q.Type,
			q.Tag != "" && !e.HasTag(q.Tag),
			q.BatchID != "" && e.BatchID != q.BatchID,
			q.FamilyHash != "" && e.FamilyHash != q.FamilyHash,
			q.Before > 0 && e.Sequence >= q.Before,
			q.IndexOnly && !e.DisplayOnIndex:
			continue
		}
		copied := *e
		out = append(out, &copied)
	}
	return out, nil
}

func (m *MemoryStorage) Find(_ context.Context, id string) (*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.entries {
		if e.UUID == id {
			copied := *e
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *MemoryStorage) Occurrences(_ context.Context, familyHash string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var n int64
	for _, e := range m.entries {
		if e.FamilyHash == familyHash {
			n++
		}
	}
	return n, nil
}

func (m *MemoryStorage) Counts(context.Context) (map[EntryType]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counts := map[EntryType]int64{}
	for _, e := range m.entries {
		counts[e.Type]++
	}
	return counts, nil
}

func (m *MemoryStorage) Prune(_ context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.entries[:0]
	for _, e := range m.entries {
		if !e.CreatedAt.Before(before) {
			kept = append(kept, e)
		}
	}
	removed := int64(len(m.entries) - len(kept))
	m.entries = kept
	return removed, nil
}

func (m *MemoryStorage) Clear(context.Context) error {
	m.mu.Lock()
	m.entries = nil
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) MonitoredTags(context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return slices.Clone(m.monitored), nil
}

func (m *MemoryStorage) Monitor(_ context.Context, tags []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range tags {
		if !slices.Contains(m.monitored, t) {
			m.monitored = append(m.monitored, t)
		}
	}
	return nil
}

func (m *MemoryStorage) Unmonitor(_ context.Context, tags []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.monitored = slices.DeleteFunc(m.monitored, func(t string) bool { return slices.Contains(tags, t) })
	return nil
}
