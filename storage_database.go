package goralens

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
)

const (
	entriesTable    = "goralens_entries"
	tagsTable       = "goralens_entries_tags"
	monitoringTable = "goralens_monitoring"
)

type DatabaseStorage struct {
	orm        orm.Orm
	connection string
}

func NewDatabaseStorage(o orm.Orm, connection string) *DatabaseStorage {
	return &DatabaseStorage{orm: o, connection: connection}
}

type entryRow struct {
	Sequence             int64     `gorm:"column:sequence;primaryKey;autoIncrement"`
	UUID                 string    `gorm:"column:uuid"`
	BatchID              string    `gorm:"column:batch_id"`
	FamilyHash           *string   `gorm:"column:family_hash"`
	ShouldDisplayOnIndex bool      `gorm:"column:should_display_on_index"`
	Type                 string    `gorm:"column:type"`
	Content              string    `gorm:"column:content"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (entryRow) TableName() string { return entriesTable }

type tagRow struct {
	EntryUUID string `gorm:"column:entry_uuid"`
	Tag       string `gorm:"column:tag"`
}

func (tagRow) TableName() string { return tagsTable }

type monitoringRow struct {
	Tag string `gorm:"column:tag"`
}

func (monitoringRow) TableName() string { return monitoringTable }

func (s *DatabaseStorage) query(ctx context.Context) orm.Query {
	o := s.orm
	if s.connection != "" {
		o = o.Connection(s.connection)
	}

	return o.WithContext(context.WithValue(ctx, skipKey{}, true)).Query()
}

func (s *DatabaseStorage) Store(ctx context.Context, entries []*Entry) error {
	if len(entries) == 0 {
		return nil
	}
	rows := make([]entryRow, 0, len(entries))
	var tags []tagRow
	families := map[string]bool{}
	for _, e := range entries {
		content, err := json.Marshal(e.Content)
		if err != nil {
			content, _ = json.Marshal(map[string]any{"error": "content could not be encoded: " + err.Error()})
		}
		row := entryRow{UUID: e.UUID, BatchID: e.BatchID, ShouldDisplayOnIndex: e.DisplayOnIndex,
			Type: string(e.Type), Content: string(content), CreatedAt: e.CreatedAt}
		if e.FamilyHash != "" {
			hash := e.FamilyHash
			row.FamilyHash = &hash
			families[hash] = true
		}
		rows = append(rows, row)
		for _, t := range e.Tags {
			tags = append(tags, tagRow{EntryUUID: e.UUID, Tag: truncate(t, 190)})
		}
	}

	for hash := range families {
		if _, err := s.query(ctx).Exec("UPDATE "+entriesTable+" SET should_display_on_index = ? WHERE family_hash = ?", false, hash); err != nil {
			return err
		}
	}
	if err := s.query(ctx).Create(&rows); err != nil {
		return err
	}
	if len(tags) > 0 {
		return s.query(ctx).Create(&tags)
	}
	return nil
}

func (s *DatabaseStorage) Get(ctx context.Context, q Query) ([]*Entry, error) {
	query := s.query(ctx).Table(entriesTable)
	if q.Type != "" {
		query = query.Where("type = ?", string(q.Type))
	}
	if q.Tag != "" {
		query = query.Where("uuid IN (SELECT entry_uuid FROM "+tagsTable+" WHERE tag = ?)", q.Tag)
	}
	if q.BatchID != "" {
		query = query.Where("batch_id = ?", q.BatchID)
	}
	if q.FamilyHash != "" {
		query = query.Where("family_hash = ?", q.FamilyHash)
	}
	if q.Before > 0 {
		query = query.Where("sequence < ?", q.Before)
	}
	if q.IndexOnly {
		query = query.Where("should_display_on_index = ?", true)
	}
	var rows []entryRow
	if err := query.OrderByDesc("sequence").Limit(q.limit()).Get(&rows); err != nil {
		return nil, err
	}
	return s.hydrate(ctx, rows)
}

func (s *DatabaseStorage) Find(ctx context.Context, id string) (*Entry, error) {
	var rows []entryRow
	if err := s.query(ctx).Table(entriesTable).Where("uuid = ?", id).Limit(1).Get(&rows); err != nil {
		return nil, err
	}
	entries, err := s.hydrate(ctx, rows)
	if err != nil || len(entries) == 0 {
		return nil, err
	}
	return entries[0], nil
}

func (s *DatabaseStorage) hydrate(ctx context.Context, rows []entryRow) ([]*Entry, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]any, len(rows))
	for i, r := range rows {
		ids[i] = r.UUID
	}
	var tags []tagRow
	if err := s.query(ctx).Table(tagsTable).WhereIn("entry_uuid", ids).Get(&tags); err != nil {
		return nil, err
	}
	byEntry := map[string][]string{}
	for _, t := range tags {
		byEntry[t.EntryUUID] = append(byEntry[t.EntryUUID], t.Tag)
	}
	entries := make([]*Entry, len(rows))
	for i, r := range rows {
		e := &Entry{Sequence: r.Sequence, UUID: r.UUID, BatchID: r.BatchID, Type: EntryType(r.Type),
			Tags: byEntry[r.UUID], CreatedAt: r.CreatedAt, DisplayOnIndex: r.ShouldDisplayOnIndex}
		if r.FamilyHash != nil {
			e.FamilyHash = *r.FamilyHash
		}
		if err := json.Unmarshal([]byte(r.Content), &e.Content); err != nil {
			e.Content = map[string]any{"error": "content could not be decoded"}
		}
		if e.Tags == nil {
			e.Tags = []string{}
		}
		entries[i] = e
	}
	return entries, nil
}

func (s *DatabaseStorage) Occurrences(ctx context.Context, familyHash string) (int64, error) {
	return s.query(ctx).Table(entriesTable).Where("family_hash = ?", familyHash).Count()
}

func (s *DatabaseStorage) Counts(ctx context.Context) (map[EntryType]int64, error) {
	var rows []struct {
		Type  string `gorm:"column:type"`
		Total int64  `gorm:"column:total"`
	}
	if err := s.query(ctx).Raw("SELECT type, COUNT(*) AS total FROM " + entriesTable + " GROUP BY type").Scan(&rows); err != nil {
		return nil, err
	}
	counts := map[EntryType]int64{}
	for _, r := range rows {
		counts[EntryType(r.Type)] = r.Total
	}
	return counts, nil
}

func (s *DatabaseStorage) Prune(ctx context.Context, before time.Time) (int64, error) {
	if _, err := s.query(ctx).Exec("DELETE FROM "+tagsTable+" WHERE entry_uuid IN (SELECT uuid FROM "+entriesTable+" WHERE created_at < ?)", before); err != nil {
		return 0, err
	}
	res, err := s.query(ctx).Exec("DELETE FROM "+entriesTable+" WHERE created_at < ?", before)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}

func (s *DatabaseStorage) Clear(ctx context.Context) error {
	if _, err := s.query(ctx).Exec("DELETE FROM " + tagsTable); err != nil {
		return err
	}
	_, err := s.query(ctx).Exec("DELETE FROM " + entriesTable)
	return err
}

func (s *DatabaseStorage) MonitoredTags(ctx context.Context) ([]string, error) {
	var rows []monitoringRow
	if err := s.query(ctx).Table(monitoringTable).Get(&rows); err != nil {
		return nil, err
	}
	tags := make([]string, len(rows))
	for i, r := range rows {
		tags[i] = r.Tag
	}
	return tags, nil
}

func (s *DatabaseStorage) Monitor(ctx context.Context, tags []string) error {
	existing, err := s.MonitoredTags(ctx)
	if err != nil {
		return err
	}
	var rows []monitoringRow
	for _, t := range tags {
		if t != "" && !contains(existing, t) {
			rows = append(rows, monitoringRow{Tag: truncate(t, 190)})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return s.query(ctx).Create(&rows)
}

func (s *DatabaseStorage) Unmonitor(ctx context.Context, tags []string) error {
	for _, t := range tags {
		if _, err := s.query(ctx).Exec("DELETE FROM "+monitoringTable+" WHERE tag = ?", t); err != nil {
			return err
		}
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
