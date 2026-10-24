package facades

import (
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	frameworkfoundation "github.com/goravel/framework/foundation"
)

func App() foundation.Application { return frameworkfoundation.App }
func Config() config.Config       { return App().MakeConfig() }
func Orm() orm.Orm                { return App().MakeOrm() }
func Log() log.Log                { return App().MakeLog() }
func Cache() cache.Cache          { return App().MakeCache() }
func Event() event.Instance       { return App().MakeEvent() }
func Queue() queue.Queue          { return App().MakeQueue() }
func Gate() access.Gate           { return App().MakeGate() }
func Route() route.Route          { return App().MakeRoute() }
