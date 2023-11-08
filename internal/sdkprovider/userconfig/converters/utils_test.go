package converters

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrillKey(t *testing.T) {
	js := `{
		"rules": {
			"mapping": [
				{
					"namespaces": ["original"],
					"namespaces_string": ["string"],
					"namespaces_object": [
						{
							"retention": "40h"
						}
					]
				}
			]
		}
	}`

	var m map[string]any
	err := json.Unmarshal([]byte(js), &m)
	require.NoError(t, err)

	cases := []struct {
		key         string
		expectOK    bool
		expectValue any
	}{
		{
			key:         "rules.0.mapping.0.namespaces",
			expectOK:    true,
			expectValue: []any{"original"},
		},
		{
			key:         "rules.0.mapping.0.namespaces_string",
			expectOK:    true,
			expectValue: []any{"string"},
		},
		{
			key:         "rules.0.mapping.0.namespaces_object",
			expectOK:    true,
			expectValue: []any{map[string]any{"retention": "40h"}},
		},
		{
			key:         "rules.0.unknown",
			expectOK:    false,
			expectValue: nil,
		},
		{
			key:         "unknown",
			expectOK:    false,
			expectValue: nil,
		},
	}

	for i, opt := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			v, ok := drillKey(m, opt.key)
			assert.Equal(t, opt.expectOK, ok)
			assert.Empty(t, cmp.Diff(opt.expectValue, v))
		})
	}
}

func TestRenameAliases(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "keeps original key",
			src:      `{"ip_filter": ["0.0.0.0/0"]}`,
			expected: `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			name:     "chooses original out if 3",
			src:      `{"ip_filter": ["0.0.0.0/0"], "ip_filter_string": [], "ip_filter_object": []}`,
			expected: `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			name:     "chooses string out if 3",
			src:      `{"ip_filter": [], "ip_filter_string": ["0.0.0.0/0"], "ip_filter_object": []}`,
			expected: `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			name:     "ignores unknown key",
			src:      `{"whatever": ["0.0.0.0/0"]}`,
			expected: `{"whatever": ["0.0.0.0/0"]}`,
		},
		{
			name:     `renames "_string" prefix`,
			src:      `{"ip_filter_string": ["0.0.0.0/0"]}`,
			expected: `{"ip_filter": ["0.0.0.0/0"]}`,
		},
		{
			name:     "renames _object prefix",
			src:      `{"ip_filter_object": [{"name": "foo"}]}`,
			expected: `{"ip_filter": [{"name": "foo"}]}`,
		},
		{
			name:     "ignores namespaces_string on the root level",
			src:      `{"namespaces_string": {"name": "foo"}}`,
			expected: `{"namespaces_string": {"name": "foo"}}`,
		},
		{
			name: "renames namespaces_string where expected",
			src: `{
				"rules": {"mapping": [{"namespaces_string": ["string"]}]}
			}`,
			expected: `{
				"rules": {"mapping": [{"namespaces": ["string"]}]}
			}`,
		},
		{
			name: "renames namespaces_object where expected",
			src: `{
				"rules": {"mapping": [{"namespaces_object": [{"name": "foo"}]}]}
			}`,
			expected: `{
				"rules": {"mapping": [{"namespaces": [{"name": "foo"}]}]}
			}`,
		},
	}

	reSpaces := regexp.MustCompile(`\s+`)
	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			var m map[string]any
			err := json.Unmarshal([]byte(opt.src), &m)
			require.NoError(t, err)

			renameAliases(m)
			b, err := json.Marshal(&m)
			require.NoError(t, err)
			assert.Empty(t, cmp.Diff(reSpaces.ReplaceAllString(opt.expected, ""), string(b)))
		})
	}
}
