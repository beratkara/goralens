package bootstrap

import (
	"github.com/beratkara/goralens"
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/cache"
	"github.com/goravel/framework/contracts/database/schema"
	contractsevent "github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/event"
	frameworkfoundation "github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/log"
	"github.com/goravel/framework/mail"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/route"
	"github.com/goravel/framework/schedule"
	"github.com/goravel/framework/validation"
	"github.com/goravel/framework/view"
	"github.com/goravel/gin"
	"github.com/goravel/sqlite"

	"github.com/beratkara/goralens/examples/basic/config"
	"github.com/beratkara/goralens/examples/basic/database/migrations"
	"github.com/beratkara/goralens/examples/basic/routes"
)

func Boot() foundation.Application {
	return frameworkfoundation.Setup().
		WithConfig(config.Boot).
		WithProviders(Providers).
		WithMigrations(func() []schema.Migration {
			return []schema.Migration{goralens.Migration{}, migrations.CreateUsersTable{}}
		}).
		WithEvents(func() map[contractsevent.Event][]contractsevent.Listener { return routes.Events() }).
		WithJobs(func() []contractsqueue.Job { return routes.Jobs() }).
		WithRouting(routes.Web).
		Create()
}

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&log.ServiceProvider{},
		&cache.ServiceProvider{},
		&http.ServiceProvider{},
		&validation.ServiceProvider{},
		&view.ServiceProvider{},
		&route.ServiceProvider{},
		&gin.ServiceProvider{},
		&database.ServiceProvider{},
		&sqlite.ServiceProvider{},
		&auth.ServiceProvider{},
		&queue.ServiceProvider{},
		&event.ServiceProvider{},
		&mail.ServiceProvider{},
		&schedule.ServiceProvider{},
		&goralens.ServiceProvider{},
	}
}
