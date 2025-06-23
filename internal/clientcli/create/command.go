package create

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/google/subcommands"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/clientcli"
	"github.com/minio/sha256-simd"
	"go.uber.org/multierr"
)

var ErrDownloadFailed = errors.New("download failed")

type ClientInterface interface {
	Snapshot(context.Context, api.SnapshotOptions) (*api.SnapshotResult, error)
	Download(context.Context, api.DownloadOptions) (*api.DownloadResult, error)
	DownloadStatus(context.Context, api.DownloadStatusOptions) (*api.DownloadStatus, error)
}

type Command struct {
	outputPath string
	format     string
	skipHead   bool
}

func (*Command) Name() string {
	return "create"
}

func (*Command) Synopsis() string {
	return `Take and download a snapshot.`
}

func (c *Command) Usage() string {
	return ``
}

func (c *Command) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.outputPath, "output", "",
		`Path to file for downloaded archive. "-" for standard output. Defaults to filename from remote side.`)
	fs.StringVar(&c.format, "format", api.ArchiveTar.Name(),
		fmt.Sprintf(`Archive format to request. One of %q.`, api.ArchiveFormatAll))
	fs.BoolVar(&c.skipHead, "skip_head", false,
		"Skip data present in the head block.")
}

func (c *Command) execute(ctx context.Context, cl ClientInterface) (err error) {
	output, err := clientcli.NewOutputFile(c.outputPath)
	if err != nil {
		return err
	}

	defer multierr.AppendInvoke(&err, multierr.Close(output))

	snapshot, err := cl.Snapshot(ctx, api.SnapshotOptions{
		SkipHead: c.skipHead,
	})
	if err != nil {
		return err
	}

	digestw := sha256.New()

	download, err := cl.Download(ctx, api.DownloadOptions{
		SnapshotName: snapshot.Name,
		Format:       api.ArchiveFormat(c.format),
		BodyWriter: func(result api.DownloadResult) (io.Writer, error) {
			w, err := output.Open(result.Filename)

			if namer, ok := w.(interface{ Name() string }); ok && err == nil {
				log.Printf("Writing snapshot archive to %s", namer.Name())
			}

			return io.MultiWriter(w, digestw), err
		},
	})
	if err != nil {
		return err
	}

	status, err := cl.DownloadStatus(ctx, api.DownloadStatusOptions{
		ID: download.ID,
	})
	if err != nil {
		return err
	}

	if sf := status.Finished; sf == nil {
		return fmt.Errorf("download not finished: %+v", status)
	} else if sf.ErrorText != nil {
		return fmt.Errorf("%w: %s", ErrDownloadFailed, *sf.ErrorText)
	} else if !sf.Success {
		return ErrDownloadFailed
	} else if want := hex.EncodeToString(digestw.Sum(nil)); sf.Sha256Hex != want {
		return fmt.Errorf("%w: SHA256 checksum mismatch (got %s, want %s)", ErrDownloadFailed, sf.Sha256Hex, want)
	}

	return nil
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
