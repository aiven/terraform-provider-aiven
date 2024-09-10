package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientOpts(t *testing.T) {
	cases := []struct {
		name      string
		opts      []ClientOpt
		envToken  string
		expectErr error
		expect    *clientOpts
	}{
		{
			name:      "empty options",
			expectErr: errTokenRequired,
		},
		{
			name:     "env token",
			envToken: "foo",
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "0.11+compatible",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/0.11+compatible/dev",
			},
		},
		{
			name: "custom token",
			opts: []ClientOpt{
				TokenOpt("opt token"),
			},
			expect: &clientOpts{
				token:        "opt token",
				tfVersion:    "0.11+compatible",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/0.11+compatible/dev",
			},
		},
		{
			name:     "custom version number",
			envToken: "foo",
			opts: []ClientOpt{
				TFVersionOpt("bar"),
			},
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "bar",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/bar/dev",
			},
		},
		{
			name:     "custom build number",
			envToken: "bar",
			opts: []ClientOpt{
				BuildVersionOpt("baz"),
			},
			expect: &clientOpts{
				token:        "bar",
				tfVersion:    "0.11+compatible",
				buildVersion: "baz",
				userAgent:    "terraform-provider-aiven/0.11+compatible/baz",
			},
		},
		{
			name: "custom all",
			opts: []ClientOpt{
				BuildVersionOpt("baz"),
				TFVersionOpt("bar"),
				TokenOpt("foo"),
			},
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "bar",
				buildVersion: "baz",
				userAgent:    "terraform-provider-aiven/bar/baz",
			},
		},
	}

	for _, o := range cases {
		t.Run(o.name, func(t *testing.T) {
			t.Setenv("AIVEN_TOKEN", o.envToken) // must not expose a real token in logs
			actual, err := newClientOpts(o.opts...)
			assert.Equal(t, o.expectErr, err)
			assert.Equal(t, o.expect, actual)
		})
	}
}
