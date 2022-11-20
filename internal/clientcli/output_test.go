package clientcli

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/internal/testutils"
)

func TestOutputFile(t *testing.T) {
	defer testutils.LogOutput(t, io.Discard)()
	defer testutils.Chdir(t, t.TempDir())()

	if err := os.WriteFile("./exists.txt", nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tmpdir := t.TempDir()

	for _, tc := range []struct {
		name    string
		path    string
		wantErr error
		want    io.Writer

		hint        string
		wantOpenErr error
	}{
		{
			name: "stdout",
			path: "-",
			want: os.Stdout,
		},
		{
			name: "file",
			path: filepath.Join(tmpdir, "output.txt"),
		},
		{
			name: "remote name",
			hint: "remote.txt",
		},
		{
			name: "remote name with path",
			hint: "../../../insecure/../../escape.txt",
		},
		{
			name:        "remote name exists",
			hint:        "exists.txt",
			wantOpenErr: os.ErrExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := NewOutputFile(tc.path)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if !(tc.want == nil || o.w == tc.want) {
				t.Errorf("Got unexpected writer %v, want %v", o.w, tc.want)
			} else if o.w == nil && tc.path != "" {
				t.Errorf("Want writer for path %q, got %v", tc.path, o.w)
			}

			w, err := o.Open(tc.hint)

			if diff := cmp.Diff(tc.wantOpenErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if err == nil && w != os.Stdout {
				if _, err := io.WriteString(w, "Test content"); err != nil {
					t.Errorf("WriteString(%v) failed: %v", w, err)
				}
			}

			if err := o.Close(); err != nil {
				t.Errorf("Close() failed: %v", err)
			}
		})
	}
}
