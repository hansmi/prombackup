package prune

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/google/subcommands"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/clientcli"
)

type ClientInterface interface {
	Prune(context.Context, api.PruneOptions) (*api.PruneResult, error)
}

type Command struct {
	keepWithin time.Duration
}

func (*Command) Name() string {
	return "prune"
}

func (*Command) Synopsis() string {
	return `Remove snapshots.`
}

func (c *Command) Usage() string {
	return ``
}

func (c *Command) SetFlags(fs *flag.FlagSet) {
	fs.DurationVar(&c.keepWithin, "keep_within", time.Hour,
		"Keep all snapshots within this time interval.")
}

func (c *Command) execute(ctx context.Context, cl ClientInterface) error {
	_, err := cl.Prune(ctx, api.PruneOptions{
		KeepWithin: c.keepWithin,
	})

	return err
}

func (c *Command) Execute(ctx context.Context, fs *flag.FlagSet, args ...any) subcommands.ExitStatus {
	r := args[0].(*clientcli.Runtime)

	if fs.NArg() != 0 {
		fs.Usage()
		return subcommands.ExitUsageError
	}

	if err := r.WithClient(func(cl api.Interface) error {
		return c.execute(ctx, cl)
	}); err != nil {
		log.Printf("Error: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
