package schemautil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSimilar(t *testing.T) {
	candidates := []string{
		"startup-2", "startup-4", "startup-8",
		"business-4", "business-8", "business-16",
		"premium-896", "premium-1792",
	}

	tests := []struct {
		target         string
		maxSuggestions int
		expected       []string
	}{
		{
			target:         "Startup-4",
			maxSuggestions: 3,
			expected:       []string{"startup-4"},
		},
		{
			target:         "startup-3",
			maxSuggestions: 2,
			expected:       []string{"startup-2", "startup-4"},
		},
		{
			target:         "buisiness-5",
			maxSuggestions: 3,
			expected:       []string{"business-4", "business-8", "business-16"},
		},
		{
			target:         "premium-900",
			maxSuggestions: 1,
			expected:       []string{"premium-896"},
		},
		{
			target:         "BUSINESS-4",
			maxSuggestions: 3,
			expected:       []string{"business-4"},
		},
		{
			target:         "nonexistent-plan",
			maxSuggestions: 2,
			expected:       []string{},
		},
	}

	for _, test := range tests {
		result := findSimilar(test.target, candidates, test.maxSuggestions)
		assert.Equal(t, test.expected, result)
	}
}
