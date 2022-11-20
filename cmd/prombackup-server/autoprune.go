package main

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/hansmi/prombackup/internal/pruner"
)

type autopruner struct {
	interval time.Duration
	opts     pruner.Options
}

func (p *autopruner) run(ctx context.Context) {
	delay := time.Duration(float32(p.interval) / 10)
	for {
		// Randomize delay
		delay = time.Duration(math.Max(float64(time.Second), float64(delay)*(0.9+(0.2*rand.Float64()))))

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}

		if err := pruner.Prune(ctx, p.opts); err != nil {
			p.opts.Logger.Printf("Pruning failed: %v", err)
		}

		delay = p.interval
	}
}
