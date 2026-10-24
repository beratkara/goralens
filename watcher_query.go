package goralens

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	pluginName  = "goralens"
	startKey    = "goralens:start"
	modelAction = "goralens:model_action"
)

type GormPlugin struct {
	Connection string
}

func (GormPlugin) Name() string { return pluginName }

func (p GormPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	register := []struct {
		before, after func() error
	}{
		{func() error { return cb.Create().Before("gorm:create").Register("goralens:before_create", before) },
			func() error {
				return cb.Create().After("gorm:create").Register("goralens:after_create", p.after("created"))
			}},
		{func() error { return cb.Query().Before("gorm:query").Register("goralens:before_query", before) },
			func() error { return cb.Query().After("gorm:query").Register("goralens:after_query", p.after("")) }},
		{func() error { return cb.Update().Before("gorm:update").Register("goralens:before_update", before) },
			func() error {
				return cb.Update().After("gorm:update").Register("goralens:after_update", p.after("updated"))
			}},
		{func() error { return cb.Delete().Before("gorm:delete").Register("goralens:before_delete", before) },
			func() error {
				return cb.Delete().After("gorm:delete").Register("goralens:after_delete", p.after("deleted"))
			}},
		{func() error { return cb.Row().Before("gorm:row").Register("goralens:before_row", before) },
			func() error { return cb.Row().After("gorm:row").Register("goralens:after_row", p.after("")) }},
		{func() error { return cb.Raw().Before("gorm:raw").Register("goralens:before_raw", before) },
			func() error { return cb.Raw().After("gorm:raw").Register("goralens:after_raw", p.after("")) }},
	}
	for _, r := range register {
		if err := r.before(); err != nil {
			return err
		}
		if err := r.after(); err != nil {
			return err
		}
	}
	return nil
}

func before(db *gorm.DB) {
	db.InstanceSet(startKey, time.Now())
}

func (p GormPlugin) after(action string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		r := Instance()
		if r == nil || db.Statement == nil || skipped(db.Statement.Context) {
			return
		}
		ctx := db.Statement.Context

		if r.config.watching(TypeQuery) && strings.TrimSpace(db.Statement.SQL.String()) != "" {
			r.Record(ctx, p.queryEntry(r, db))
		}
		if action != "" && db.Error == nil && db.Statement.Schema != nil && r.config.watching(TypeModel) {
			for _, e := range modelEntries(db, action) {
				if changes, ok := e.Content["changes"]; ok {
					e.Content["changes"] = redactMap(changes, r.config.HiddenRequestParameters)
				}
				r.Record(ctx, e)
			}
		}
	}
}

func (p GormPlugin) queryEntry(r *Recorder, db *gorm.DB) *Entry {
	var elapsed time.Duration
	if v, ok := db.InstanceGet(startKey); ok {
		if start, ok := v.(time.Time); ok {
			elapsed = time.Since(start)
		}
	}
	stmt := db.Statement
	sql := stmt.SQL.String()
	content := map[string]any{
		"connection": p.Connection,
		"sql":        db.Dialector.Explain(sql, stmt.Vars...),
		"raw_sql":    sql,
		"bindings":   exportValue(stmt.Vars),
		"time":       milliseconds(elapsed),
		"rows":       db.RowsAffected,
		"slow":       elapsed >= r.config.SlowQuery,
	}
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		content["error"] = db.Error.Error()
	}
	if f, ok := origin(captureStack(2)); ok {
		content["file"], content["line"] = f.File, f.Line
	}
	e := newEntry(TypeQuery, content)
	if elapsed >= r.config.SlowQuery {
		e.AddTags("slow")
	}
	return e
}

func modelEntries(db *gorm.DB, action string) []*Entry {
	stmt := db.Statement
	value := reflect.Indirect(stmt.ReflectValue)
	var models []reflect.Value
	switch value.Kind() {
	case reflect.Struct:
		models = append(models, value)
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len() && i < 50; i++ {
			models = append(models, reflect.Indirect(value.Index(i)))
		}
	}
	var entries []*Entry
	for _, m := range models {
		if m.Kind() != reflect.Struct {
			continue
		}
		key := primaryKey(stmt.Schema, stmt, m)
		content := map[string]any{"action": action, "model": stmt.Schema.Name, "table": stmt.Table, "key": key}
		if action != "deleted" {
			content["changes"] = changedFields(stmt, m, action)
		}
		e := newEntry(TypeModel, content)
		e.AddTags(fmt.Sprintf("%s:%v", stmt.Schema.Name, key))
		entries = append(entries, e)
	}
	if len(entries) == 0 && action != "created" {

		entries = append(entries, newEntry(TypeModel, map[string]any{
			"action": action, "model": stmt.Schema.Name, "table": stmt.Table, "key": nil,
			"changes": exportValue(stmt.Dest), "rows": db.RowsAffected,
		}))
	}
	return entries
}

func primaryKey(s *schema.Schema, stmt *gorm.Statement, m reflect.Value) any {
	if s.PrioritizedPrimaryField == nil {
		return nil
	}
	v, zero := s.PrioritizedPrimaryField.ValueOf(stmt.Context, m)
	if zero {
		return nil
	}
	return v
}

func changedFields(stmt *gorm.Statement, m reflect.Value, action string) map[string]any {
	out := map[string]any{}
	if action == "updated" {
		if dest, ok := stmt.Dest.(map[string]any); ok {
			for k, v := range dest {
				out[k] = exportValue(v)
			}
			return out
		}
	}
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		if action == "updated" && len(stmt.Selects) > 0 && !contains(stmt.Selects, f.DBName) && !contains(stmt.Selects, f.Name) {
			continue
		}
		v, zero := f.ValueOf(stmt.Context, m)
		if zero && action == "updated" {
			continue
		}
		out[f.DBName] = exportValue(v)
	}
	return out
}
