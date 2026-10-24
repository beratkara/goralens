package goralens

import (
	"testing"
	"time"

	"github.com/goravel/framework/contracts/foundation"
)

type listedWatcher struct{}

func (listedWatcher) Name() string                          { return "listed" }
func (listedWatcher) Register(foundation.Application) error { return nil }

func TestOnlyListedWatchersRun(t *testing.T) {
	c := defaultConfig()
	applyWatchers(c, []Watcher{
		RequestWatcher{SizeLimit: 8, IgnoreStatusCodes: []int{404}},
		&QueryWatcher{SlowMs: 250, Connections: []string{"pgsql"}},
		LogWatcher{Level: "warning"},
		ExceptionWatcher{IgnoreErrorLogs: true},
		listedWatcher{},
	})
	for _, typ := range EntryTypes {
		want := typ == TypeRequest || typ == TypeQuery || typ == TypeLog || typ == TypeException
		if c.Watchers[typ] != want {
			t.Errorf("%s enabled = %v, want %v", typ, c.Watchers[typ], want)
		}
	}
	if c.ResponseSizeLimit != 8 || len(c.IgnoreStatusCodes) != 1 || c.SlowQuery != 250*time.Millisecond ||
		c.QueryConnections[0] != "pgsql" || c.LogLevel != "warning" || c.ExceptionsFromLogs {
		t.Fatalf("watcher options not applied: %+v", c)
	}
	if len(c.CustomWatchers) != 1 || c.CustomWatchers[0].Name() != "listed" {
		t.Fatalf("custom watchers: %+v", c.CustomWatchers)
	}
}

func TestDefaultWatchersKeepDefaults(t *testing.T) {
	c := defaultConfig()
	applyWatchers(c, DefaultWatchers())
	for _, typ := range EntryTypes {
		if !c.Watchers[typ] {
			t.Errorf("%s should be on", typ)
		}
	}
	if c.ResponseSizeLimit != 64 || c.SlowQuery != 100*time.Millisecond || c.LogLevel != "debug" || !c.ExceptionsFromLogs {
		t.Fatalf("defaults changed: %+v", c)
	}
}
