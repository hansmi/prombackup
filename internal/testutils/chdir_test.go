package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChdir(t *testing.T) {
	tmpdir := t.TempDir()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %v", err)
	}

	restore := Chdir(t, tmpdir)

	if wd, err := os.Getwd(); err != nil {
		t.Errorf("Getwd() failed: %v", err)
	} else if got, want := filepath.Clean(wd), filepath.Clean(tmpdir); got != want {
		t.Errorf("Chdir() changed to %s, not %s", got, want)
	}

	restore()

	if wd, err := os.Getwd(); err != nil {
		t.Errorf("Getwd() failed: %v", err)
	} else if got, want := filepath.Clean(wd), filepath.Clean(original); got != want {
		t.Errorf("Restore call changed to %s, not %s", got, want)
	}
}
