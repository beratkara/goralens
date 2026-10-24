package goralens

import (
	"errors"

	"github.com/goravel/framework/contracts/database/schema"
)

type Migration struct{}

func (Migration) Signature() string {
	return "20260101000000_create_goralens_tables"
}

func (Migration) Connection() string {
	if app == nil {
		return ""
	}
	return loadConfig(app.MakeConfig()).Connection
}

func (Migration) Up() error {
	s, err := migrationSchema()
	if err != nil {
		return err
	}
	if !s.HasTable(entriesTable) {
		if err := s.Create(entriesTable, func(table schema.Blueprint) {
			table.BigIncrements("sequence")
			table.Uuid("uuid")
			table.Uuid("batch_id")
			table.String("family_hash").Nullable()
			table.Boolean("should_display_on_index").Default(true)
			table.String("type", 20)
			table.LongText("content")
			table.DateTime("created_at").Nullable()

			table.Unique("uuid")
			table.Index("batch_id")
			table.Index("family_hash")
			table.Index("created_at")
			table.Index("type", "should_display_on_index")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable(tagsTable) {
		if err := s.Create(tagsTable, func(table schema.Blueprint) {
			table.Uuid("entry_uuid")
			table.String("tag", 190)

			table.Primary("entry_uuid", "tag")
			table.Index("tag")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable(monitoringTable) {
		if err := s.Create(monitoringTable, func(table schema.Blueprint) {
			table.String("tag", 190)
			table.Primary("tag")
		}); err != nil {
			return err
		}
	}
	return nil
}

func (Migration) Down() error {
	s, err := migrationSchema()
	if err != nil {
		return err
	}
	for _, t := range []string{tagsTable, monitoringTable, entriesTable} {
		if err := s.DropIfExists(t); err != nil {
			return err
		}
	}
	return nil
}

func migrationSchema() (schema.Schema, error) {
	if app == nil {
		return nil, errors.New("goralens: the service provider is not registered")
	}
	s := app.MakeSchema()
	if s == nil {
		return nil, errors.New("goralens: the schema facade is not available")
	}
	return s, nil
}
