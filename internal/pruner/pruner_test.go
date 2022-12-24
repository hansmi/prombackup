package pruner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		want    snapshotInfo
		wantErr error
	}{
		{
			name: "20221115T232711Z-7ef1661077569104",
			want: snapshotInfo{
				Timestamp: time.Date(2022, 11, 15, 23, 27, 11, 0, time.UTC),
				Name:      "20221115T232711Z-7ef1661077569104",
			},
		},
		{
			name:    "",
			wantErr: errInvalidName,
		},
		{
			name:    "bad-more-bad",
			wantErr: errInvalidName,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseName(tc.name)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseName() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSelectForDeletion(t *testing.T) {
	for _, tc := range []struct {
		name  string
		opts  Options
		input []snapshotInfo
		want  []snapshotInfo
	}{
		{name: "empty"},
		{
			name:  "zero timestamp",
			input: []snapshotInfo{{}, {}},
		},
		{
			name: "all",
			opts: Options{
				nowFunc: func() time.Time {
					return time.Date(2022, 1, 1, 13, 0, 0, 0, time.UTC)
				},
			},
			input: []snapshotInfo{
				{
					Name:      "1980",
					Timestamp: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:      "13:00",
					Timestamp: time.Date(2022, 1, 1, 13, 0, 0, 0, time.UTC),
				},
				{
					Name:      "future",
					Timestamp: time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: []snapshotInfo{
				{
					Name:      "1980",
					Timestamp: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:      "13:00",
					Timestamp: time.Date(2022, 1, 1, 13, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "keep within",
			opts: Options{
				KeepWithin: 24 * time.Hour,
				nowFunc: func() time.Time {
					return time.Date(2020, 8, 15, 0, 0, 0, 0, time.UTC)
				},
			},
			input: []snapshotInfo{
				{
					Name:      "1990",
					Timestamp: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:      "13:00",
					Timestamp: time.Date(2020, 8, 15, 13, 0, 0, 0, time.UTC),
				},
				{
					Name:      "future",
					Timestamp: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:      "previous day",
					Timestamp: time.Date(2020, 8, 13, 23, 45, 0, 0, time.UTC),
				},
			},
			want: []snapshotInfo{
				{
					Name:      "1990",
					Timestamp: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:      "previous day",
					Timestamp: time.Date(2020, 8, 13, 23, 45, 0, 0, time.UTC),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.opts.selectForDeletion(tc.input)

			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("selectForDeletion() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrune(t *testing.T) {
	tmpdirAll := t.TempDir()
	tmpdirSelective := t.TempDir()
	tmpdirCheck := t.TempDir()

	for _, i := range []struct {
		root    string
		subdirs []string
	}{
		{
			root: tmpdirAll,
			subdirs: []string{
				"20221110T230339Z-7aac2bc56f45ca8e",
				"bad",
				"20221115T232711Z-7ef1661077569104",
			},
		},
		{
			root: tmpdirSelective,
			subdirs: []string{
				"20200101T000000Z-a",
				"20211110T230339Z-b",
				"20221110T230339+0000-c",
				"20221115T232711Z-d",
			},
		},
		{
			root: tmpdirCheck,
			subdirs: []string{
				"20181018T000000Z-a",
				"20191019T000000Z-b",
				"20201020T000000Z-c",
			},
		},
	} {
		for _, j := range i.subdirs {
			if err := os.Mkdir(filepath.Join(i.root, j), 0o777); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, tc := range []struct {
		name          string
		opts          Options
		wantRemaining []string
		wantErr       error
	}{
		{
			name: "empty",
			opts: Options{
				Root: t.TempDir(),
			},
		},
		{
			name: "all",
			opts: Options{
				Root: tmpdirAll,
			},
			wantRemaining: []string{
				"bad",
			},
		},
		{
			name: "selective",
			opts: Options{
				Root:       tmpdirSelective,
				KeepWithin: time.Hour,
				nowFunc: func() time.Time {
					return time.Date(2022, 1, 1, 13, 0, 0, 0, time.UTC)
				},
			},
			wantRemaining: []string{
				"20221110T230339+0000-c",
				"20221115T232711Z-d",
			},
		},
		{
			name: "pre remove check",
			opts: Options{
				Root:       tmpdirCheck,
				KeepWithin: time.Hour,
				PreRemoveCheck: func(name string) error {
					if name == "20191019T000000Z-b" {
						return nil
					}

					return fmt.Errorf("fake error for %s", name)
				},
				nowFunc: func() time.Time {
					return time.Date(2022, 1, 1, 13, 0, 0, 0, time.UTC)
				},
			},
			wantRemaining: []string{
				"20181018T000000Z-a",
				"20201020T000000Z-c",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Prune(context.Background(), tc.opts)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			var remaining []string

			if entries, err := os.ReadDir(tc.opts.Root); err != nil {
				t.Errorf("ReadDir() failed: %v", err)
			} else {
				for _, entry := range entries {
					remaining = append(remaining, entry.Name())
				}
			}

			sort.Strings(remaining)

			if diff := cmp.Diff(tc.wantRemaining, remaining, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Remaining entries diff (-want +got):\n%s", diff)
			}
		})
	}
}
