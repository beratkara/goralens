package goralens

import (
	"errors"
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/binding"
	contractscache "github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/console"
	contractsevent "github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/mail"
	"gorm.io/gorm"
)

const Binding = "goralens"

var app foundation.Application

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{Binding},
		Dependencies: []string{
			binding.Config, binding.Log, binding.Orm, binding.Schema, binding.Route, binding.Artisan,
			binding.Cache, binding.Event, binding.Queue, binding.Mail, binding.Gate, binding.Http,
			binding.Schedule,
		},
		ProvideFor: []string{},
	}
}

func (r *ServiceProvider) Register(a foundation.Application) {
	app = a
	a.Singleton(Binding, func(foundation.Application) (any, error) {
		if rec := Instance(); rec != nil {
			return rec, nil
		}
		return nil, errors.New("goralens is disabled")
	})
}

func (r *ServiceProvider) Boot(a foundation.Application) {
	cfg := loadConfig(a.MakeConfig())
	storageFor := func() Storage { return newStorage(a, cfg) }
	a.Commands([]console.Command{
		&PruneCommand{storage: storageFor},
		&ClearCommand{storage: storageFor},
	})
	if !cfg.Enabled {
		return
	}
	rec := NewRecorder(cfg, storageFor())
	SetRecorder(rec)

	if route := a.MakeRoute(); route != nil {
		route.GlobalMiddleware(NewMiddleware(rec))
	}
	if cfg.watching(TypeQuery) || cfg.watching(TypeModel) {
		installGormPlugin(a, cfg)
	}
	if cfg.watching(TypeGate) && hasBinding(a, binding.Gate) {
		WatchGate(a.MakeGate())
	}
	if cfg.watching(TypeCache) {
		decorate(a, binding.Cache, func(v any) any { return WrapCache(v.(contractscache.Cache)) })
	}
	if cfg.watching(TypeEvent) {
		decorate(a, binding.Event, func(v any) any { return WrapEvent(v.(contractsevent.Instance)) })
	}
	if cfg.watching(TypeJob) {
		decorate(a, binding.Queue, func(v any) any { return WrapQueue(v.(contractsqueue.Queue)) })
	}
	if cfg.watching(TypeClientRequest) {
		decorate(a, binding.Http, func(v any) any { return WrapHttp(v.(client.Factory)) })
	}
	if cfg.watching(TypeMail) {
		watchMail(a)
	}
	decorate(a, binding.Artisan, func(v any) any { return WrapArtisan(v.(console.Artisan)) })
	if err := bootCustomWatchers(a, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if cfg.watching(TypeSchedule) {
		decorate(a, binding.Schedule, func(v any) any {
			return WrapSchedule(v.(contractsschedule.Schedule), a.MakeArtisan())
		})
	}
}

func newStorage(a foundation.Application, cfg *Config) Storage {
	if cfg.Driver == "memory" {
		return NewMemoryStorage(0)
	}
	if !hasBinding(a, binding.Orm) {
		return NewMemoryStorage(0)
	}
	o := a.MakeOrm()
	if o == nil {
		return NewMemoryStorage(0)
	}
	return NewDatabaseStorage(o, cfg.Connection)
}

func installGormPlugin(a foundation.Application, cfg *Config) {
	if !hasBinding(a, binding.Orm) {
		return
	}
	o := a.MakeOrm()
	if o == nil {
		return
	}
	connections := cfg.QueryConnections
	if len(connections) == 0 {
		connections = []string{""}
	}
	for _, name := range connections {
		conn := o
		if name != "" {
			conn = o.Connection(name)
		}
		q, ok := conn.Query().(interface{ Instance() *gorm.DB })
		if !ok || q.Instance() == nil {
			continue
		}
		label := name
		if label == "" {
			label = o.Name()
		}
		if err := q.Instance().Use(GormPlugin{Connection: label}); err != nil && !errors.Is(err, gorm.ErrRegistered) {
			fmt.Fprintf(os.Stderr, "goralens: installing the query watcher on %q failed: %v\n", label, err)
		}
	}
}

func watchMail(a foundation.Application) {
	if !hasBinding(a, binding.Mail) {
		return
	}
	if _, ok := a.MakeMail().(*mail.Application); !ok {
		return
	}
	a.Bind(binding.Mail, func(a foundation.Application) (any, error) {
		m, err := mail.NewApplication(a.MakeConfig(), a.MakeQueue())
		if err != nil {
			return nil, err
		}
		return WrapMail(m), nil
	})
}

func decorate(a foundation.Application, key string, wrap func(any) any) {
	if !hasBinding(a, key) {
		return
	}
	original, err := a.Make(key)
	if err != nil || original == nil {
		return
	}
	a.Instance(key, wrap(original))
	a.Fresh(key)
}

func hasBinding(a foundation.Application, key string) bool {
	for _, b := range a.Bindings() {
		if b == key {
			return true
		}
	}
	return false
}
