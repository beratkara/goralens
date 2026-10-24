package goralens

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type PruneCommand struct {
	storage func() Storage
}

func (c *PruneCommand) Signature() string { return "goralens:prune" }
func (c *PruneCommand) Description() string {
	return "Delete Goralens entries older than the given number of hours"
}

func (c *PruneCommand) Extend() command.Extend {
	return command.Extend{
		Category: "goralens",
		Flags: []command.Flag{
			&command.IntFlag{Name: "hours", Value: 24, Usage: "Keep entries created within this many hours"},
		},
	}
}

func (c *PruneCommand) Handle(ctx console.Context) error {
	storage := c.storage()
	if storage == nil {
		ctx.Error("Goralens storage is not available.")
		return nil
	}
	hours := ctx.OptionInt("hours")
	if hours < 0 {
		hours = 0
	}
	removed, err := storage.Prune(context.WithValue(context.Background(), skipKey{}, true), time.Now().Add(-time.Duration(hours)*time.Hour))
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Info(fmt.Sprintf("%d entries pruned.", removed))
	return nil
}

type ClearCommand struct {
	storage func() Storage
}

func (c *ClearCommand) Signature() string      { return "goralens:clear" }
func (c *ClearCommand) Description() string    { return "Delete all Goralens entries" }
func (c *ClearCommand) Extend() command.Extend { return command.Extend{Category: "goralens"} }

func (c *ClearCommand) Handle(ctx console.Context) error {
	storage := c.storage()
	if storage == nil {
		ctx.Error("Goralens storage is not available.")
		return nil
	}
	if err := storage.Clear(context.WithValue(context.Background(), skipKey{}, true)); err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Info("Goralens entries cleared.")
	return nil
}
