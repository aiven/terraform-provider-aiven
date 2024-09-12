package schemautil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsRedactedCreds(t *testing.T) {
	cases := []struct {
		name     string
		hash     map[string]any
		expected error
	}{
		{
			name:     "contains redacted",
			hash:     map[string]any{"password": "<redacted>"},
			expected: errContainsRedactedCreds,
		},
		{
			name:     "contains invalid redacted",
			hash:     map[string]any{"password": "<REDACTED>"},
			expected: nil,
		},
		{
			name:     "does not contain redacted",
			hash:     map[string]any{"password": "123"},
			expected: nil,
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			err := ContainsRedactedCreds(opt.hash)
			assert.Equal(t, err, opt.expected)
		})
	}
}
