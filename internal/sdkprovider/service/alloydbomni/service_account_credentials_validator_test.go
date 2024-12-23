package alloydbomni

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestValidateServiceAccountCredentials(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected diag.Diagnostics
	}{
		{
			name: "valid",
			input: `{
				"private_key_id": "0",
				"private_key": "1",
				"client_email": "2",
				"client_id": "3"
			}`,
			expected: nil,
		},
		{
			name:  "invalid, empty",
			input: `{}`,
			expected: diag.Diagnostics{
				{Summary: "(root): private_key_id is required"},
				{Summary: "(root): private_key is required"},
				{Summary: "(root): client_email is required"},
				{Summary: "(root): client_id is required"},
			},
		},
		{
			name: "missing private_key_id",
			input: `{
				"private_key": "1",
				"client_email": "2",
				"client_id": "3"
			}`,
			expected: diag.Diagnostics{{Summary: "(root): private_key_id is required"}},
		},
		{
			name: "invalid type client_id",
			input: `{
				"private_key_id": "0",
				"private_key": "1",
				"client_email": "2",
				"client_id": 3
			}`,
			expected: diag.Diagnostics{{Summary: "client_id: Invalid type. Expected: string, given: integer"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validateServiceAccountCredentials(tc.input, nil)
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				assert.Empty(t, diff)
			}
		})
	}
}
