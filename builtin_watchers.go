package goralens

import (
	"time"

	"github.com/goravel/framework/contracts/foundation"
)

type builtinWatcher interface {
	Watcher
	apply(c *Config)
}

type RequestWatcher struct {
	SizeLimit         int
	IgnoreStatusCodes []int
}

func (RequestWatcher) Name() string                          { return string(TypeRequest) }
func (RequestWatcher) Register(foundation.Application) error { return nil }
func (w RequestWatcher) apply(c *Config) {
	if w.SizeLimit > 0 {
		c.ResponseSizeLimit = w.SizeLimit
	}
	c.IgnoreStatusCodes = w.IgnoreStatusCodes
}

type ClientRequestWatcher struct{}

func (ClientRequestWatcher) Name() string                          { return string(TypeClientRequest) }
func (ClientRequestWatcher) Register(foundation.Application) error { return nil }
func (ClientRequestWatcher) apply(*Config)                         {}

type CommandWatcher struct{}

func (CommandWatcher) Name() string                          { return string(TypeCommand) }
func (CommandWatcher) Register(foundation.Application) error { return nil }
func (CommandWatcher) apply(*Config)                         {}

type ScheduleWatcher struct{}

func (ScheduleWatcher) Name() string                          { return string(TypeSchedule) }
func (ScheduleWatcher) Register(foundation.Application) error { return nil }
func (ScheduleWatcher) apply(*Config)                         {}

type JobWatcher struct{}

func (JobWatcher) Name() string                          { return string(TypeJob) }
func (JobWatcher) Register(foundation.Application) error { return nil }
func (JobWatcher) apply(*Config)                         {}

type EventWatcher struct{}

func (EventWatcher) Name() string                          { return string(TypeEvent) }
func (EventWatcher) Register(foundation.Application) error { return nil }
func (EventWatcher) apply(*Config)                         {}

type GateWatcher struct {
	IgnoreAbilities []string
}

func (GateWatcher) Name() string                          { return string(TypeGate) }
func (GateWatcher) Register(foundation.Application) error { return nil }
func (w GateWatcher) apply(c *Config)                     { c.IgnoreAbilities = w.IgnoreAbilities }

type MailWatcher struct{}

func (MailWatcher) Name() string                          { return string(TypeMail) }
func (MailWatcher) Register(foundation.Application) error { return nil }
func (MailWatcher) apply(*Config)                         {}

type ExceptionWatcher struct {
	IgnoreErrorLogs bool
}

func (ExceptionWatcher) Name() string                          { return string(TypeException) }
func (ExceptionWatcher) Register(foundation.Application) error { return nil }
func (w ExceptionWatcher) apply(c *Config)                     { c.ExceptionsFromLogs = !w.IgnoreErrorLogs }

type LogWatcher struct {
	Level string
}

func (LogWatcher) Name() string                          { return string(TypeLog) }
func (LogWatcher) Register(foundation.Application) error { return nil }
func (w LogWatcher) apply(c *Config) {
	if w.Level != "" {
		c.LogLevel = w.Level
	}
}

type DumpWatcher struct{}

func (DumpWatcher) Name() string                          { return string(TypeDump) }
func (DumpWatcher) Register(foundation.Application) error { return nil }
func (DumpWatcher) apply(*Config)                         {}

type QueryWatcher struct {
	SlowMs      int
	Connections []string
}

func (QueryWatcher) Name() string                          { return string(TypeQuery) }
func (QueryWatcher) Register(foundation.Application) error { return nil }
func (w QueryWatcher) apply(c *Config) {
	if w.SlowMs > 0 {
		c.SlowQuery = time.Duration(w.SlowMs) * time.Millisecond
	}
	c.QueryConnections = w.Connections
}

type ModelWatcher struct{}

func (ModelWatcher) Name() string                          { return string(TypeModel) }
func (ModelWatcher) Register(foundation.Application) error { return nil }
func (ModelWatcher) apply(*Config)                         {}

type CacheWatcher struct {
	HiddenKeys []string
}

func (CacheWatcher) Name() string                          { return string(TypeCache) }
func (CacheWatcher) Register(foundation.Application) error { return nil }
func (w CacheWatcher) apply(c *Config)                     { c.HiddenCacheKeys = w.HiddenKeys }

type RedisWatcher struct{}

func (RedisWatcher) Name() string                          { return string(TypeRedis) }
func (RedisWatcher) Register(foundation.Application) error { return nil }
func (RedisWatcher) apply(*Config)                         {}

func DefaultWatchers() []Watcher {
	return []Watcher{
		RequestWatcher{}, ClientRequestWatcher{}, CommandWatcher{}, ScheduleWatcher{}, JobWatcher{},
		EventWatcher{}, GateWatcher{}, MailWatcher{}, ExceptionWatcher{}, LogWatcher{}, DumpWatcher{},
		QueryWatcher{}, ModelWatcher{}, CacheWatcher{}, RedisWatcher{},
	}
}

func applyWatchers(c *Config, list []Watcher) {
	for _, t := range EntryTypes {
		c.Watchers[t] = false
	}
	c.CustomWatchers = nil
	for _, w := range list {
		if w == nil {
			continue
		}
		if b, ok := w.(builtinWatcher); ok {
			c.Watchers[EntryType(b.Name())] = true
			b.apply(c)
			continue
		}
		c.CustomWatchers = append(c.CustomWatchers, w)
	}
}
