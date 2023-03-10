package vpc

import (
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/stretchr/testify/require"
)

func TestGetVPC(t *testing.T) {
	differentCloudName := []*aiven.VPC{
		{ProjectVPCID: "foo", CloudName: "aws"},
		{ProjectVPCID: "bar", CloudName: "azure"},
	}

	duplicateCloudName := []*aiven.VPC{
		{ProjectVPCID: "foo", CloudName: "azure"},
		{ProjectVPCID: "bar", CloudName: "azure"},
	}

	cases := []struct {
		name           string
		inputVPCList   []*aiven.VPC
		inputVPCID     string
		inputCloudName string
		expectVPC      *aiven.VPC
		expectErr      error
	}{
		{
			name:           `same cloudName, gets by vpcID "foo"`,
			inputVPCList:   duplicateCloudName,
			inputVPCID:     "foo",
			inputCloudName: "",
			expectVPC:      &aiven.VPC{ProjectVPCID: "foo", CloudName: "azure"},
			expectErr:      nil,
		},
		{
			name:           `same cloudName, gets by vpcID "bar"`,
			inputVPCList:   duplicateCloudName,
			inputVPCID:     "bar",
			inputCloudName: "",
			expectVPC:      &aiven.VPC{ProjectVPCID: "bar", CloudName: "azure"},
			expectErr:      nil,
		},
		{
			name:           `different cloudName, gets by cloudName "azure"`,
			inputVPCList:   differentCloudName,
			inputVPCID:     "",
			inputCloudName: "azure",
			expectVPC:      &aiven.VPC{ProjectVPCID: "bar", CloudName: "azure"},
			expectErr:      nil,
		},
		{
			name:           `same cloudName, gets err`,
			inputVPCList:   duplicateCloudName,
			inputVPCID:     "",
			inputCloudName: "azure",
			expectVPC:      nil,
			expectErr:      fmt.Errorf(`multiple project VPC with cloud_name "azure", use vpc_id instead`),
		},
		{
			name:           `invalid input XNOR`,
			inputVPCList:   duplicateCloudName,
			inputVPCID:     "foo",
			inputCloudName: "azure",
			expectVPC:      nil,
			expectErr:      fmt.Errorf(`provide exactly one: vpc_id or cloud_name`),
		},
		{
			name:           `invalid input XNOR, both empty`,
			inputVPCList:   duplicateCloudName,
			inputVPCID:     "",
			inputCloudName: "",
			expectVPC:      nil,
			expectErr:      fmt.Errorf(`provide exactly one: vpc_id or cloud_name`),
		},
		{
			name:           `nothing found for cloudName "lol"`,
			inputVPCList:   differentCloudName,
			inputVPCID:     "",
			inputCloudName: "lol",
			expectVPC:      nil,
			expectErr:      fmt.Errorf(`not found project VPC with cloud_name "lol"`),
		},
		{
			name:           `nothing found for empty slice"`,
			inputVPCList:   nil,
			inputVPCID:     "",
			inputCloudName: "lol",
			expectVPC:      nil,
			expectErr:      fmt.Errorf(`not found project VPC with cloud_name "lol"`),
		},
	}

	for _, o := range cases {
		t.Run(o.name, func(t *testing.T) {
			vpc, err := getVPC(o.inputVPCList, o.inputVPCID, o.inputCloudName)
			require.Equal(t, err, o.expectErr)
			require.Equal(t, vpc, o.expectVPC)

			// Never returns both nil, also xnor
			require.NotEqual(t, err == nil, vpc == nil)
		})
	}
}
