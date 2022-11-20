package clientcli

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/internal/ref"
)

const envVarName = "pb_test_var"

func TestGetenvBool(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    *string
		fallback bool
		want     bool
		wantErr  error
	}{
		{name: "unset"},
		{
			name:  "empty",
			value: ref.Ref(""),
		},
		{
			name:  "true",
			value: ref.Ref("1"),
			want:  true,
		},
		{
			name:  "false",
			value: ref.Ref("0"),
			want:  false,
		},
		{
			name:     "fallback",
			fallback: true,
			want:     true,
		},
		{
			name:    "error",
			value:   ref.Ref("nope"),
			wantErr: strconv.ErrSyntax,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv(envVarName)

			if tc.value != nil {
				os.Setenv(envVarName, *tc.value)
			}

			got, err := GetenvBool(envVarName, tc.fallback)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetenvBool diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetenvDuration(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    *string
		fallback time.Duration
		want     time.Duration
		wantErr  error
	}{
		{name: "unset"},
		{
			name:  "empty",
			value: ref.Ref(""),
		},
		{
			name:  "1h3m",
			value: ref.Ref("1h3m"),
			want:  time.Hour + 3*time.Minute,
		},
		{
			name:     "fallback",
			fallback: 13 * time.Hour,
			want:     13 * time.Hour,
		},
		{
			name:    "error",
			value:   ref.Ref("nope"),
			wantErr: cmpopts.AnyError,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv(envVarName)

			if tc.value != nil {
				os.Setenv(envVarName, *tc.value)
			}

			got, err := GetenvDuration(envVarName, tc.fallback)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetenvDuration diff (-want +got):\n%s", diff)
			}
		})
	}
}
