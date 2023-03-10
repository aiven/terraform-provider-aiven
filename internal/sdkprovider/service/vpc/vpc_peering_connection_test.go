package vpc

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateVPCID(t *testing.T) {
	type args struct {
		i interface{}
		k string
	}
	tests := []struct {
		name         string
		args         args
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name:         "basic",
			args:         args{"project/vpc", "vpc_id"},
			wantWarnings: nil,
			wantErrors:   nil,
		},
		{
			name:         "wrong id",
			args:         args{"wrong", "vpc_id"},
			wantWarnings: nil,
			wantErrors:   []error{errors.New("invalid vpc_id, expected <project_name>/<vpc_id>")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := validateVPCID(tt.args.i, tt.args.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("validateVPCID() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("validateVPCID() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func Test_parsePeerVPCID(t *testing.T) {
	region := "peerRegion"
	cases := []struct {
		name     string
		source   string
		expected *peeringVPCID
		err      error
	}{
		{
			name:     "empty string",
			source:   "",
			expected: nil,
			err:      fmt.Errorf("expected unix path-like string with 4-5 chunks, got 1"),
		},
		{
			name:     "too many chunks",
			source:   "projectName/vpcID/peerCloudAccount/peerVPC/peerRegion/foo/bar/baz/egg",
			expected: nil,
			err:      fmt.Errorf("expected unix path-like string with 4-5 chunks, got 9"),
		},
		{
			name:   "full 5 chunks",
			source: "projectName/vpcID/peerCloudAccount/peerVPC/peerRegion",
			expected: &peeringVPCID{
				projectName:      "projectName",
				vpcID:            "vpcID",
				peerCloudAccount: "peerCloudAccount",
				peerVPC:          "peerVPC",
				peerRegion:       &region,
			},
		},
		{
			name:   "only 4 chunks",
			source: "projectName/vpcID/peerCloudAccount/peerVPC",
			expected: &peeringVPCID{
				projectName:      "projectName",
				vpcID:            "vpcID",
				peerCloudAccount: "peerCloudAccount",
				peerVPC:          "peerVPC",
				peerRegion:       nil,
			},
		},
		{
			name:     "only 3 chunks",
			source:   "projectName/vpcID/peerCloudAccount",
			expected: nil,
			err:      fmt.Errorf("expected unix path-like string with 4-5 chunks, got 3"),
		},
		{
			name:     "only 2 chunks",
			source:   "projectName/vpcID",
			expected: nil,
			err:      fmt.Errorf("expected unix path-like string with 4-5 chunks, got 2"),
		},
		{
			name:     "only 1 chunk",
			source:   "projectName",
			expected: nil,
			err:      fmt.Errorf("expected unix path-like string with 4-5 chunks, got 1"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := parsePeerVPCID(c.source)
			assert.Equal(t, c.err, err)
			assert.Equal(t, c.expected, actual)
		})
	}
}
