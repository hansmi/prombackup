package snapshotstream

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hansmi/prombackup/api"
	"github.com/klauspost/compress/gzip"
	"go.uber.org/multierr"
)

var ErrArchiveFormat = errors.New("unknown archive format")
var ErrNotFound = errors.New("snapshot not found")
var ErrInvalid = errors.New("snapshot invalid")

type Options struct {
	Name   string
	Root   fs.FS
	Format api.ArchiveFormat
}

type Stream struct {
	ContentType string
	Filename    string

	id     string
	name   string
	root   fs.FS
	format api.ArchiveFormat

	mu     sync.Mutex
	status api.DownloadStatus
}

func New(opts Options) (*Stream, error) {
	if opts.Name == "" {
		return nil, errors.New("name is required")
	}

	s := &Stream{
		id:     newID(),
		name:   opts.Name,
		root:   opts.Root,
		format: opts.Format,
	}

	switch f := opts.Format; f {
	case api.ArchiveTar, api.ArchiveTarGzip:
		s.ContentType = f.ContentType()
		s.Filename = filepath.Base(s.name) + f.FileExtension()
	default:
		return nil, fmt.Errorf("%w: %s", ErrArchiveFormat, f.Name())
	}

	s.status.ID = s.id
	s.status.SnapshotName = s.name

	if fi, err := fs.Stat(s.root, "."); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, s.name)
		}

		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrInvalid, s.name)
	}

	return s, nil
}

func (s *Stream) ID() string {
	return s.id
}

func (s *Stream) Name() string {
	return s.name
}

func (s *Stream) Status() api.DownloadStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.status
}

func (s *Stream) writeTo(w io.Writer) (err error) {
	var archiveWriter io.Writer
	var compressionFlush func() error

	switch s.format {
	case api.ArchiveTar:
		archiveWriter = w

	case api.ArchiveTarGzip:
		gzipWriter := gzip.NewWriter(w)
		gzipWriter.Header.Name = strings.TrimSuffix(filepath.Base(s.Filename), filepath.Ext(s.Filename))
		gzipWriter.Header.Comment = fmt.Sprintf("Prometheus snapshot %s", s.name)
		gzipWriter.Header.ModTime = time.Now()

		defer multierr.AppendInvoke(&err, multierr.Close(gzipWriter))

		archiveWriter = gzipWriter
		compressionFlush = gzipWriter.Flush

	default:
		return ErrArchiveFormat
	}

	a := newTarArchiver(archiveWriter, compressionFlush)

	defer multierr.AppendInvoke(&err, multierr.Close(a))

	return archiveDir(s.root, s.name, a)
}

func (s *Stream) WriteTo(w io.Writer) error {
	digestw := sha256.New()

	err := s.writeTo(io.MultiWriter(w, digestw))

	sf := api.DownloadStatusFinished{
		Success: (err == nil),
	}

	if err == nil {
		sf.Sha256Hex = hex.EncodeToString(digestw.Sum(nil))
	} else {
		msg := err.Error()
		sf.ErrorText = &msg
	}

	s.mu.Lock()
	s.status.Finished = &sf
	s.mu.Unlock()

	return err
}
