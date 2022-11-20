package snapshotstream

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var errTest = errors.New("test error")

type tarEntry struct {
	name    string
	content string
}

func checkTarContents(t *testing.T, r io.Reader, want []tarEntry) {
	var got []tarEntry

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("Next() failed: %v", err)
		}

		e := tarEntry{
			name: hdr.Name,
		}

		if content, err := io.ReadAll(tr); err != nil {
			t.Errorf("ReadAll() failed: %v", err)
		} else {
			e.content = string(content)
		}

		got = append(got, e)
	}

	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmp.AllowUnexported(tarEntry{})); diff != "" {
		t.Errorf("Archive contents diff (-want +got):\n%s", diff)
	}
}

func TestTarArchiver(t *testing.T) {
	for _, tc := range []struct {
		name           string
		entries        []archiveEntry
		wantFileErrors error
		want           []tarEntry
	}{
		{name: "empty"},
		{
			name: "dirs",
			entries: []archiveEntry{
				{
					name: "rootdir",
					mode: fs.ModeDir,
				},
				{
					name: "rootdir/sub1",
					mode: fs.ModeDir,
				},
			},
			want: []tarEntry{
				{name: "rootdir"},
				{name: "rootdir/sub1"},
			},
		},
		{
			name: "mixed",
			entries: []archiveEntry{
				{
					name: "test.txt",
					data: "test content",
				},
				{
					name: "dir",
					mode: fs.ModeDir,
				},
				{
					name: "dir/hello.txt",
					data: "world",
				},
			},
			want: []tarEntry{
				{
					name:    "test.txt",
					content: "test content",
				},
				{name: "dir"},
				{
					name:    "dir/hello.txt",
					content: "world",
				},
			},
		},
		{
			name: "unsupported socket",
			entries: []archiveEntry{
				{
					name: "socket",
					mode: fs.ModeSocket,
				},
			},
			wantFileErrors: errUnsupportedType,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			a := newTarArchiver(&buf, nil)

			archiverTest{
				entries:        tc.entries,
				wantFileErrors: tc.wantFileErrors,
			}.do(t, a)

			if err := a.Close(); err != nil {
				t.Errorf("Close() failed: %v", err)
			}

			checkTarContents(t, &buf, tc.want)
		})
	}
}

func TestTarArchiverLarge(t *testing.T) {
	for _, tc := range []struct {
		name           string
		root           fs.FS
		flushErr       error
		wantErr        error
		wantFlushCount int
	}{
		{
			name: "empty",
			root: &fstest.MapFS{},
		},
		{
			name: "large files",
			root: &fstest.MapFS{
				"small": {
					Data: []byte("hello"),
				},
				"large": {
					Data: bytes.Repeat([]byte("test"), 1024*1024),
				},
				"small2": {
					Data: []byte("world"),
				},
				"more": {
					Data: bytes.Repeat([]byte("xyz"), 1024*1024),
				},
			},
			wantFlushCount: 2,
		},
		{
			name:     "flush error",
			flushErr: errTest,
			root: &fstest.MapFS{
				"empty": {},
				"trigger": {
					Data: bytes.Repeat([]byte("bad"), 1024*1024),
				},
			},
			wantErr:        errTest,
			wantFlushCount: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			var flushCount int

			a := newTarArchiver(&buf, func() error {
				flushCount++
				return tc.flushErr
			})

			err := archiveDir(tc.root, ".", a)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("archiveDir() error diff (-want +got):\n%s", diff)
			}

			if flushCount != tc.wantFlushCount {
				t.Errorf("Flush count is %d, want %d", flushCount, tc.wantFlushCount)
			}
		})
	}
}
