package clientcli

import (
	"io"
	"os"
	"path/filepath"
)

type OutputFile struct {
	path string

	w         io.WriteCloser
	needClose bool
}

func NewOutputFile(path string) (*OutputFile, error) {
	switch path {
	case "":
		return &OutputFile{}, nil

	case "-":
		return &OutputFile{
			w: os.Stdout,
		}, nil
	}

	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &OutputFile{
		w:         fh,
		needClose: true,
	}, nil
}

func (f *OutputFile) Open(hint string) (io.Writer, error) {
	if f.w != nil {
		return f.w, nil
	}

	path := filepath.Base(hint)

	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0o666)
	if err != nil {
		return nil, err
	}

	f.w = fh
	f.needClose = true

	return f.w, nil
}

func (f *OutputFile) Close() error {
	if f.w != nil && f.needClose {
		return f.w.Close()
	}

	return nil
}
