package testutils

import (
	"os"
	"testing"
)

func Chdir(t *testing.T, path string) func() {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %v", err)
	}

	if err := os.Chdir(path); err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("Restoring working directory: %v", err)
		}
	}
}
