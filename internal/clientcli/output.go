package clientcli

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
)

const outputBufferSize = 1024 * 1024

type OutputFile struct {
	path string

	w     io.Writer
	close func() error
}

func NewOutputFile(path string) (*OutputFile, error) {
	f := &OutputFile{}

	switch path {
	case "":
		// Use remote filename

	case "-":
		f.w = os.Stdout

	default:
		fh, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		f.setup(fh)
	}

	return f, nil
}

func (f *OutputFile) setup(w io.WriteCloser) {
	if f.w != nil {
		return
	}

	buf := bufio.NewWriterSize(w, outputBufferSize)

	f.w = buf
	f.close = func() (err error) {
		defer multierr.AppendInvoke(&err, multierr.Close(w))

		return buf.Flush()
	}
}

// Open returns a writer. If a path was given to [NewOutputFile] the hint is
// unused. Otherwise it's used to create a new file.
func (f *OutputFile) Open(hint string) (io.Writer, error) {
	if f.w != nil {
		return f.w, nil
	}

	path := filepath.Base(hint)

	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0o666)
	if err != nil {
		return nil, err
	}

	f.setup(fh)

	return f.w, nil
}

func (f *OutputFile) Close() error {
	if f.close != nil {
		return f.close()
	}

	return nil
}
