package snapshotstream

import (
	"io"
	"io/fs"
	"path/filepath"

	"go.uber.org/multierr"
)

type openFunc func() (io.ReadCloser, error)

type archiver interface {
	io.Closer
	Append(string, fs.DirEntry, openFunc) error
	FileErrors() error
}

func archiveDir(root fs.FS, base string, a archiver) error {
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Directory walk failed
			return err
		}

		return a.Append(filepath.Join(base, path), d, func() (io.ReadCloser, error) {
			return root.Open(path)
		})
	})

	multierr.AppendInto(&err, a.FileErrors())

	return err
}
