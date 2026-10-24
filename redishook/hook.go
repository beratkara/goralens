package redishook

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/beratkara/goralens"
	"github.com/redis/go-redis/v9"
)

type Hook struct {
	connection string
}

func New(connection string) *Hook {
	return &Hook{connection: connection}
}

func (h *Hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		goralens.RecordRedis(ctx, h.connection, Command(cmd), time.Since(start), realError(err))
		return err
	}
}

func (h *Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)

		elapsed := time.Since(start) / time.Duration(max(len(cmds), 1))
		for _, cmd := range cmds {
			goralens.RecordRedis(ctx, h.connection, Command(cmd), elapsed, realError(cmd.Err()))
		}
		return err
	}
}

func Command(cmd redis.Cmder) string {
	parts := make([]string, len(cmd.Args()))
	for i, arg := range cmd.Args() {
		s := fmt.Sprint(arg)
		if strings.ContainsAny(s, " \t\n\"") || s == "" {
			s = fmt.Sprintf("%q", s)
		}
		parts[i] = s
	}
	return strings.Join(parts, " ")
}

func realError(err error) error {
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}
