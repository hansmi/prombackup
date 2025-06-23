package snapshotstream

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
	"github.com/klauspost/compress/zstd"
)

func TestStream(t *testing.T) {
	for _, tc := range []struct {
		name             string
		opts             Options
		wantNewErr       error
		wantContentType  string
		wantFilename     string
		wantStatusBefore api.DownloadStatus
		wantStatusAfter  api.DownloadStatus
		wantGzipHeader   gzip.Header
		wantErr          error
		want             []tarEntry
	}{
		{
			name: "empty",
			opts: Options{
				Name:   "empty",
				Root:   &fstest.MapFS{},
				Format: api.ArchiveTar,
			},
			wantContentType: "application/x-tar",
			wantFilename:    "empty.tar",
			wantStatusAfter: api.DownloadStatus{
				SnapshotName: "empty",
				Finished: &api.DownloadStatusFinished{
					Success: true,
				},
			},
			want: []tarEntry{
				{name: "empty"},
			},
		},
		{
			name: "unsupported file type",
			opts: Options{
				Name: "unsupported",
				Root: &fstest.MapFS{
					"aaa": {
						Data: []byte("content"),
					},
					"socket": {
						Mode: fs.ModeSocket,
					},
				},
				Format: api.ArchiveTar,
			},
			wantContentType: "application/x-tar",
			wantFilename:    "unsupported.tar",
			wantStatusAfter: api.DownloadStatus{
				SnapshotName: "unsupported",
				Finished: &api.DownloadStatusFinished{
					Success: false,
				},
			},
			wantErr: errUnsupportedType,
			want: []tarEntry{
				{name: "unsupported"},
				{
					name:    "unsupported/aaa",
					content: "content",
				},
			},
		},
		{
			name: "tar+gzip",
			opts: Options{
				Name: "archive93c2",
				Root: &fstest.MapFS{
					"file": {
						Data: []byte("hello world"),
					},
				},
				Format: api.ArchiveTarGzip,
			},
			wantContentType: "application/gzip",
			wantFilename:    "archive93c2.tar.gz",
			wantStatusAfter: api.DownloadStatus{
				SnapshotName: "archive93c2",
				Finished: &api.DownloadStatusFinished{
					Success: true,
				},
			},
			wantGzipHeader: gzip.Header{
				Comment: "Prometheus snapshot archive93c2",
				Name:    "archive93c2.tar",
			},
			want: []tarEntry{
				{name: "archive93c2"},
				{name: "archive93c2/file", content: "hello world"},
			},
		},
		{
			name: "tar+zstd",
			opts: Options{
				Name: "archive51b2fb",
				Root: &fstest.MapFS{
					"file": {
						Data: []byte("hello world"),
					},
				},
				Format: api.ArchiveTarZstd,
			},
			wantContentType: "application/zstd",
			wantFilename:    "archive51b2fb.tar.zst",
			wantStatusAfter: api.DownloadStatus{
				SnapshotName: "archive51b2fb",
				Finished: &api.DownloadStatusFinished{
					Success: true,
				},
			},
			want: []tarEntry{
				{name: "archive51b2fb"},
				{name: "archive51b2fb/file", content: "hello world"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(tc.opts)

			if diff := cmp.Diff(tc.wantNewErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("New() error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantContentType, s.ContentType); diff != "" {
				t.Errorf("Content type diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantFilename, s.Filename); diff != "" {
				t.Errorf("Filename diff (-want +got):\n%s", diff)
			}

			tc.wantStatusBefore.ID = s.ID()
			tc.wantStatusBefore.SnapshotName = s.Name()
			tc.wantStatusAfter.ID = tc.wantStatusBefore.ID

			opts := []cmp.Option{
				cmpopts.EquateErrors(),
				cmpopts.EquateEmpty(),
				cmpopts.IgnoreFields(api.DownloadStatusFinished{}, "ErrorText", "Sha256Hex"),
			}

			if diff := cmp.Diff(tc.wantStatusBefore, s.Status(), opts...); diff != "" {
				t.Errorf("Status() diff (-want +got):\n%s", diff)
			}

			var buf bytes.Buffer

			err = s.WriteArchive(&buf)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantStatusAfter, s.Status(), opts...); diff != "" {
				t.Errorf("Status() diff (-want +got):\n%s", diff)
			}

			var tarReader io.Reader = &buf

			switch tc.opts.Format {
			case api.ArchiveTar:
			case api.ArchiveTarGzip:
				gzReader, err := gzip.NewReader(tarReader)
				if err != nil {
					t.Errorf("gzip.NewReader() failed: %v", err)
				}

				if diff := cmp.Diff(tc.wantGzipHeader, gzReader.Header, cmpopts.IgnoreFields(gzip.Header{}, "OS", "ModTime")); diff != "" {
					t.Errorf("Gzip header diff (-want +got):\n%s", diff)
				}

				tarReader = gzReader

			case api.ArchiveTarZstd:
				zstdReader, err := zstd.NewReader(tarReader)
				if err != nil {
					t.Errorf("zstd.NewReader() failed: %v", err)
				}

				tarReader = zstdReader

			default:
				t.Errorf("Unhandled format %q", tc.opts.Format)
			}

			checkTarContents(t, tarReader, tc.want)
		})
	}
}
