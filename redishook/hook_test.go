package redishook

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestCommandFormatting(t *testing.T) {
	cmd := redis.NewStringCmd(context.Background(), "set", "greeting", "hello world")
	if got := Command(cmd); got != `set greeting "hello world"` {
		t.Fatalf("got %s", got)
	}
}

func TestProcessHookPassesThrough(t *testing.T) {
	called := false
	hook := New("default").ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
		called = true
		return redis.Nil
	})
	if err := hook(context.Background(), redis.NewStringCmd(context.Background(), "get", "missing")); err != redis.Nil {
		t.Fatalf("error changed: %v", err)
	}
	if !called {
		t.Fatal("next hook not called")
	}
}
