package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/beratkara/goralens/examples/basic/app/facades"
)

type CreateUsersTable struct{}

func (CreateUsersTable) Signature() string { return "20260101000001_create_users_table" }

func (CreateUsersTable) Up() error {
	s := facades.App().MakeSchema()
	if s.HasTable("users") {
		return nil
	}
	return s.Create("users", func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("email")
		table.String("password")
		table.Timestamps()
	})
}

func (CreateUsersTable) Down() error {
	return facades.App().MakeSchema().DropIfExists("users")
}
