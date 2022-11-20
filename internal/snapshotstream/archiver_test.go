package snapshotstream

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type archiveEntry struct {
	name    string
	mode    fs.FileMode
	data    string
	wantErr error
}

func (e *archiveEntry) args(t *testing.T) (fs.DirEntry, openFunc) {
	t.Helper()

	fs := fstest.MapFS{
		e.name: {
			Mode: e.mode,
			Data: []byte(e.data),
		},
	}

	entries, err := fs.ReadDir(filepath.Dir(e.name))
	if err != nil {
		t.Fatal(err)
	}

	return entries[0], func() (io.ReadCloser, error) {
		return fs.Open(e.name)
	}
}

type archiverTest struct {
	entries        []archiveEntry
	wantFileErrors error
}

func (tc archiverTest) do(t *testing.T, a archiver) {
	t.Helper()

	for _, e := range tc.entries {
		d, open := e.args(t)

		err := a.Append(e.name, d, open)

		if diff := cmp.Diff(e.wantErr, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("Append() error diff (-want +got):\n%s", diff)
		}
	}

	if diff := cmp.Diff(tc.wantFileErrors, a.FileErrors(), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("FileErrors() error diff (-want +got):\n%s", diff)
	}

	if err := a.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestArchiveDir(t *testing.T) {
	for _, tc := range []struct {
		name string
		root fs.FS
		base string
		want []tarEntry
	}{
		{
			name: "empty",
			root: &fstest.MapFS{},
			want: []tarEntry{
				{name: "."},
			},
		},
		{
			name: "nested",
			root: &fstest.MapFS{
				"dir": {
					Mode: fs.ModeDir,
				},
				"dir/sub": {
					Mode: fs.ModeDir,
				},
				"root.txt": {
					Data: []byte("root"),
				},
				"dir/aaa.txt": {
					Data: []byte("aaa"),
				},
			},
			base: "base",
			want: []tarEntry{
				{name: "base"},
				{name: "base/dir"},
				{
					name:    "base/dir/aaa.txt",
					content: "aaa",
				},
				{name: "base/dir/sub"},
				{
					name:    "base/root.txt",
					content: "root",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			a := newTarArchiver(&buf, nil)

			err := archiveDir(tc.root, tc.base, a)
			if err != nil {
				t.Errorf("archiveDir() failed: %v", err)
			}

			if err := a.Close(); err != nil {
				t.Errorf("Close() failed: %v", err)
			}

			checkTarContents(t, &buf, tc.want)
		})
	}
}
