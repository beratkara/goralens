package goralens

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goravel/framework/contracts/http"
)

type User struct {
	ID         any            `json:"id"`
	Name       string         `json:"name,omitempty"`
	Email      string         `json:"email,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (u *User) Field(name string) (string, bool) {
	if u == nil {
		return "", false
	}
	switch strings.ToLower(name) {
	case "id":
		if u.ID == nil {
			return "", false
		}
		return fmt.Sprint(u.ID), true
	case "name":
		return u.Name, u.Name != ""
	case "email":
		return u.Email, u.Email != ""
	}
	for k, v := range u.Attributes {
		if strings.EqualFold(k, name) && v != nil {
			return fmt.Sprint(v), true
		}
	}
	return "", false
}

type hooks struct {
	auth   func(ctx http.Context) bool
	filter func(e *Entry) bool
	tag    func(e *Entry) []string
	user   func(ctx http.Context) *User
}

var (
	hooksMu sync.RWMutex
	hookSet hooks

	instance atomic.Pointer[Recorder]
)

func currentHooks() hooks {
	hooksMu.RLock()
	defer hooksMu.RUnlock()
	return hookSet
}

func Auth(fn func(ctx http.Context) bool) {
	hooksMu.Lock()
	hookSet.auth = fn
	hooksMu.Unlock()
}

func Filter(fn func(e *Entry) bool) {
	hooksMu.Lock()
	hookSet.filter = fn
	hooksMu.Unlock()
}

func Tag(fn func(e *Entry) []string) {
	hooksMu.Lock()
	hookSet.tag = fn
	hooksMu.Unlock()
}

func ResolveUser(fn func(ctx http.Context) *User) {
	hooksMu.Lock()
	hookSet.user = fn
	hooksMu.Unlock()
}

func SetRecorder(r *Recorder) {
	instance.Store(r)
}

func Instance() *Recorder {
	return instance.Load()
}

func Record(ctx context.Context, e *Entry) {
	Instance().Record(ctx, e)
}

func milliseconds(d time.Duration) float64 {
	return float64(d.Microseconds()/10) / 100
}
