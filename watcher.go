package goralens

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
)

type Watcher interface {
	Name() string
	Register(app foundation.Application) error
}

type Column struct {
	Label string `json:"label"`
	Field string `json:"field"`
}

type WatcherInfo struct {
	Title   string   `json:"title"`
	Group   string   `json:"group"`
	Hint    string   `json:"hint"`
	Columns []Column `json:"columns"`
}

type Describer interface {
	Describe() WatcherInfo
}

type customWatcher struct {
	Type EntryType `json:"type"`
	WatcherInfo
}

var (
	watchersMu       sync.RWMutex
	registeredExtras []Watcher
	activeCustom     []customWatcher
)

var watcherName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

func RegisterWatcher(w Watcher) {
	watchersMu.Lock()
	registeredExtras = append(registeredExtras, w)
	watchersMu.Unlock()
}

func NewEntry(t EntryType, content map[string]any) *Entry {
	return newEntry(t, content)
}

func customWatchers() []customWatcher {
	watchersMu.RLock()
	defer watchersMu.RUnlock()
	return append([]customWatcher(nil), activeCustom...)
}

func bootCustomWatchers(a foundation.Application, cfg *Config) error {
	watchersMu.Lock()
	list := append(append([]Watcher(nil), cfg.CustomWatchers...), registeredExtras...)
	activeCustom = nil
	watchersMu.Unlock()

	seen := map[EntryType]bool{}
	for _, t := range EntryTypes {
		seen[t] = true
	}
	var active []customWatcher
	for _, w := range list {
		if w == nil {
			continue
		}
		t := EntryType(w.Name())
		if !watcherName.MatchString(string(t)) {
			return fmt.Errorf("goralens: watcher name %q must be lowercase letters, digits or underscores", t)
		}
		if seen[t] {
			return fmt.Errorf("goralens: watcher %q is registered twice or clashes with a built-in watcher", t)
		}
		seen[t] = true
		cfg.Watchers[t] = true
		if err := w.Register(a); err != nil {
			delete(cfg.Watchers, t)
			return fmt.Errorf("goralens: registering watcher %q: %w", t, err)
		}
		info := WatcherInfo{Title: string(t), Group: "Custom"}
		if d, ok := w.(Describer); ok {
			info = d.Describe()
			if info.Title == "" {
				info.Title = string(t)
			}
			if info.Group == "" {
				info.Group = "Custom"
			}
		}
		active = append(active, customWatcher{Type: t, WatcherInfo: info})
	}

	watchersMu.Lock()
	activeCustom = active
	watchersMu.Unlock()
	return nil
}
