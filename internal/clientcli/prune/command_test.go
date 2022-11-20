package prune

import (
	"context"
	"errors"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
)

var errTest = errors.New("test error")

type fakeClient struct {
	pruneResult api.PruneResult
	pruneError  error
}

func (c *fakeClient) Prune(context.Context, api.PruneOptions) (*api.PruneResult, error) {
	return &c.pruneResult, c.pruneError
}

func TestCommand(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		client  *fakeClient
		wantErr error
	}{
		{
			name:   "success",
			client: &fakeClient{},
		},
		{
			name: "error",
			client: &fakeClient{
				pruneError: errTest,
			},
			wantErr: errTest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("", flag.ContinueOnError)

			var c Command

			c.SetFlags(fs)

			if err := fs.Parse(tc.args); err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}

			err := c.execute(context.Background(), tc.client)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}
