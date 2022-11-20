package pruner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errInvalidName = errors.New("invalid snapshot name")

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}

type snapshotInfo struct {
	Timestamp time.Time
	Name      string
}

func parseName(name string) (snapshotInfo, error) {
	dashPos := strings.IndexByte(name, '-')
	if dashPos < 0 {
		return snapshotInfo{}, fmt.Errorf("%w: %s", errInvalidName, name)
	}

	ts, err := time.ParseInLocation("20060102T150405Z0700", name[:dashPos], time.UTC)
	if err != nil {
		return snapshotInfo{}, fmt.Errorf("%w: %s", errInvalidName, err.Error())
	}

	return snapshotInfo{
		Timestamp: ts,
		Name:      name,
	}, nil
}

type Options struct {
	Logger         Logger
	Root           string
	KeepWithin     time.Duration
	PreRemoveCheck func(string) error

	nowFunc func() time.Time
}

func (o *Options) selectForDeletion(snapshots []snapshotInfo) []snapshotInfo {
	if o.nowFunc == nil {
		o.nowFunc = time.Now
	}

	now := o.nowFunc()

	keep := func(ts time.Time) bool {
		if ts.IsZero() || now.Before(ts) {
			return true
		}

		return now.Before(ts.Add(o.KeepWithin))
	}

	var result []snapshotInfo

	for _, info := range snapshots {
		if !keep(info.Timestamp) {
			result = append(result, info)
		}
	}

	return result
}

func Prune(ctx context.Context, opts Options) error {
	if opts.Logger == nil {
		opts.Logger = log.New(io.Discard, "", 0)
	}

	if opts.KeepWithin < 0 {
		return fmt.Errorf("KeepWithin must be larger than or equal to zero")
	}

	entries, err := os.ReadDir(opts.Root)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var snapshots []snapshotInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := parseName(entry.Name())
		if err != nil {
			opts.Logger.Printf("Ignoring directory %s: %s", entry.Name(), err)
			continue
		}

		snapshots = append(snapshots, info)
	}

	for _, info := range opts.selectForDeletion(snapshots) {
		if opts.PreRemoveCheck != nil {
			if err := opts.PreRemoveCheck(info.Name); err != nil {
				opts.Logger.Printf("Not removing snapshot %s: %s", info.Name, err)
				continue
			}
		}

		opts.Logger.Printf("Delete snapshot %s", info.Name)

		path := filepath.Join(opts.Root, filepath.Base(info.Name))

		if err := os.RemoveAll(path); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
