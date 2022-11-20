package snapshotstream

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"go.uber.org/multierr"
)

var errUnsupportedType = errors.New("unsupported file type")

// Flush underlying writer (usually compression) before writing a file larger
// than this size in bytes.
const tarFlushMinSize = 1024 * 1024

type tarArchiver struct {
	flush   func() error
	tw      *tar.Writer
	copybuf []byte
	fileErr error
}

func newTarArchiver(w io.Writer, flush func() error) *tarArchiver {
	return &tarArchiver{
		flush:   flush,
		tw:      tar.NewWriter(w),
		copybuf: make([]byte, 1024*1024),
	}
}

func (a *tarArchiver) Close() error {
	return a.tw.Close()
}

func (a *tarArchiver) FileErrors() error {
	return a.fileErr
}

// Append writes meta information and file content to the archive. Globally
// fatal errors are returned straight away while per-file errors are collected.
// Directories and regular files are the only supported types.
func (a *tarArchiver) Append(name string, d fs.DirEntry, open openFunc) (err error) {
	hdr := tar.Header{
		Name: name,
	}

	if d.IsDir() {
		hdr.Typeflag = tar.TypeDir
		hdr.Mode = 0o755
	} else if d.Type().IsRegular() {
		hdr.Typeflag = tar.TypeReg
		hdr.Mode = 0o644
	} else {
		multierr.AppendInto(&a.fileErr, fmt.Errorf("%w: %s (%s)", errUnsupportedType, name, d.Type()))
		return nil
	}

	fi, err := d.Info()
	if err != nil {
		multierr.AppendInto(&a.fileErr, fmt.Errorf("%s: %w", name, err))
		return nil
	}

	hdr.ModTime = fi.ModTime()

	if hdr.Typeflag == tar.TypeReg {
		hdr.Size = fi.Size()
	}

	if hdr.Size > tarFlushMinSize && a.flush != nil {
		// Force-flush before writing a larger file. This will more likely
		// result in equal outputs for the same data, which is useful for rsync
		// and deduplicating backup systems.
		if err := a.flush(); err != nil {
			return err
		}
	}

	if err := a.tw.WriteHeader(&hdr); err != nil {
		return err
	}

	if hdr.Typeflag == tar.TypeReg {
		fh, err := open()
		if err != nil {
			return err
		}

		defer multierr.AppendInvoke(&err, multierr.Close(fh))

		_, err = io.CopyBuffer(a.tw, fh, a.copybuf)

		return err
	}

	return nil
}
