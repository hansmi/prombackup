package main

import (
	"context"
	"testing"
	"time"

	"github.com/hansmi/prombackup/internal/pruner"
)

func TestAutoprunerRunCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := autopruner{
		interval: time.Minute,
		opts: pruner.Options{
			Root: t.TempDir(),
		},
	}

	p.run(ctx)
}
