package goralens

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
)

func WrapArtisan(a console.Artisan) console.Artisan {
	if a == nil {
		return nil
	}
	if _, ok := a.(*watchedArtisan); ok {
		return a
	}
	return &watchedArtisan{Artisan: a}
}

type watchedArtisan struct {
	console.Artisan
}

func (w *watchedArtisan) Call(command string) error {
	return w.run(strings.Fields(command), func() error { return w.Artisan.Call(command) })
}

func (w *watchedArtisan) CallAndExit(command string) {
	err := w.Call(command)
	exit(err)
}

func (w *watchedArtisan) Run(args []string, exitIfArtisan bool) error {
	i := indexOf(args, "artisan")
	if i < 0 {
		return w.Artisan.Run(args, exitIfArtisan)
	}
	err := w.run(args[i+1:], func() error { return w.Artisan.Run(args, false) })
	if exitIfArtisan {
		exit(err)
	}
	return err
}

func (w *watchedArtisan) run(args []string, run func() error) error {
	r := Instance()
	name := "list"
	if len(args) > 0 {
		name = args[0]
	}
	if r == nil || !r.Recording() || matchesAny(name, r.config.IgnoreCommands) || strings.HasPrefix(name, "goralens:") {
		return run()
	}
	start := time.Now()
	_, end := r.JoinOrStartBatch()
	err := run()
	if r.config.watching(TypeCommand) {
		content := map[string]any{
			"command":   name,
			"arguments": args[min(1, len(args)):],
			"exit_code": 0,
			"duration":  milliseconds(time.Since(start)),
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			content["exit_code"] = 1
			content["error"] = err.Error()
		}
		e := newEntry(TypeCommand, content)
		e.CreatedAt = start
		r.Record(context.Background(), e)
	}
	end()

	r.Flush(5 * time.Second)
	return err
}

func exit(err error) {
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func indexOf(list []string, v string) int {
	for i, item := range list {
		if item == v {
			return i
		}
	}
	return -1
}
