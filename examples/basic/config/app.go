package config

import (
	"github.com/beratkara/goralens"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	ginfacades "github.com/goravel/gin/facades"
	sqlitefacades "github.com/goravel/sqlite/facades"

	"github.com/beratkara/goralens/examples/basic/app/facades"
	"github.com/beratkara/goralens/examples/basic/app/watchers"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":     "Goralens Example",
		"env":      config.Env("APP_ENV", "local"),
		"debug":    false,
		"timezone": "UTC",
		"locale":   "en",
		"key":      config.Env("APP_KEY", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	})
	config.Add("http", map[string]any{
		"default": "gin",
		"drivers": map[string]any{
			"gin": map[string]any{
				"body_limit":   4096,
				"header_limit": 4096,
				"route":        func() (route.Route, error) { return ginfacades.Route("gin"), nil },
			},
		},
		"url":             "http://localhost",
		"host":            "127.0.0.1",
		"port":            config.Env("APP_PORT", "3000"),
		"request_timeout": 10,
		"default_client":  "default",
		"clients":         map[string]any{"default": map[string]any{"timeout": "5s"}},
	})
	config.Add("database", map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"database": config.Env("DB_DATABASE", "database.sqlite"),
				"prefix":   "",
				"singular": false,
				"via":      func() (driver.Driver, error) { return sqlitefacades.Sqlite("sqlite") },
			},
		},
		"migrations": map[string]any{"table": "migrations"},
	})
	config.Add("logging", map[string]any{
		"default": "stack",
		"channels": map[string]any{
			"stack":    map[string]any{"driver": "stack", "channels": []string{"single", "goralens"}},
			"single":   map[string]any{"driver": "single", "path": "storage/logs/app.log", "level": "debug", "print": false},
			"goralens": map[string]any{"driver": "custom", "via": goralens.LogDriver()},
		},
	})
	config.Add("cache", map[string]any{
		"default": "memory",
		"stores":  map[string]any{"memory": map[string]any{"driver": "memory"}},
		"prefix":  "",
	})
	config.Add("queue", map[string]any{
		"default":     "sync",
		"connections": map[string]any{"sync": map[string]any{"driver": "sync"}},
		"failed":      map[string]any{"database": "sqlite", "table": "failed_jobs"},
	})
	config.Add("mail", map[string]any{
		"host": "", "port": 587,
		"from": map[string]any{"address": "hello@example.com", "name": "Example"},
	})
	config.Add("goralens", map[string]any{
		"enabled":           config.Env("GORALENS_ENABLED", true),
		"path":              "goralens",
		"storage":           map[string]any{"driver": config.Env("GORALENS_DRIVER", "database")},
		"token_storage_key": config.Env("GORALENS_TOKEN_STORAGE_KEY", "app_token"),
		"middleware": []goralens.DashboardMiddleware{
			goralens.Authorize(),
			func(ctx http.Context) bool { return ctx.Request().Header("X-Block") != "yes" },
		},
		"watchers": []goralens.Watcher{
			goralens.RequestWatcher{},
			goralens.ClientRequestWatcher{},
			goralens.CommandWatcher{},
			goralens.ScheduleWatcher{},
			goralens.JobWatcher{},
			goralens.EventWatcher{},
			goralens.GateWatcher{},
			goralens.MailWatcher{},
			goralens.ExceptionWatcher{},
			goralens.LogWatcher{},
			goralens.DumpWatcher{},
			goralens.QueryWatcher{SlowMs: 50},
			goralens.ModelWatcher{},
			goralens.CacheWatcher{},
			watchers.Payments{},
		},
	})
}
