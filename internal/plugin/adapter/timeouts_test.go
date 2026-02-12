// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// timeoutsTestSchema returns a minimal schema for testing getTimeoutValue with NewResourceDataFromMaps.
func timeoutsTestSchema() *Schema {
	return &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"id": {Type: SchemaTypeString},
			"timeouts": {
				Type: SchemaTypeList,
				Items: &Schema{
					Type: SchemaTypeObject,
					Properties: map[string]*Schema{
						"create":  {Type: SchemaTypeString},
						"read":    {Type: SchemaTypeString},
						"update":  {Type: SchemaTypeString},
						"delete":  {Type: SchemaTypeString},
						"default": {Type: SchemaTypeString},
					},
				},
			},
		},
	}
}

func TestGetTimeoutValue(t *testing.T) {
	t.Parallel()

	const fallbackTimeout = 5 * time.Minute
	tests := []struct {
		name            string
		timeouts        map[string]any // create, read, update, delete, default
		timeoutType     timeoutType
		want            time.Duration
		wantErrorSubstr string // if non-empty, expect error containing this substring
	}{
		{
			name:        "valid create timeout",
			timeouts:    map[string]any{"create": "30s"},
			timeoutType: timeoutCreate,
			want:        30 * time.Second,
		},
		{
			name:        "valid read timeout",
			timeouts:    map[string]any{"read": "45s"},
			timeoutType: timeoutRead,
			want:        45 * time.Second,
		},
		{
			name:        "valid update timeout",
			timeouts:    map[string]any{"update": "1m"},
			timeoutType: timeoutUpdate,
			want:        1 * time.Minute,
		},
		{
			name:        "valid delete timeout",
			timeouts:    map[string]any{"delete": "2m"},
			timeoutType: timeoutDelete,
			want:        2 * time.Minute,
		},
		{
			name:        "valid default timeout",
			timeouts:    map[string]any{"default": "1m"},
			timeoutType: timeoutCreate,
			want:        1 * time.Minute,
		},
		{
			name:        "create falls back to default value",
			timeouts:    map[string]any{"default": "90s"},
			timeoutType: timeoutCreate,
			want:        90 * time.Second,
		},
		{
			name:            "invalid duration format",
			timeouts:        map[string]any{"create": "invalid"},
			timeoutType:     timeoutCreate,
			want:            0,
			wantErrorSubstr: "failed to parse timeout",
		},
		{
			name:        "uses default when timeout not specified",
			timeouts:    map[string]any{"update": "45s"},
			timeoutType: timeoutCreate,
			want:        fallbackTimeout,
		},
		{
			name:        "empty timeouts uses fallback",
			timeouts:    nil,
			timeoutType: timeoutCreate,
			want:        fallbackTimeout,
		},
		{
			name:        "complex duration format",
			timeouts:    map[string]any{"create": "1h2m3s"},
			timeoutType: timeoutCreate,
			want:        1*time.Hour + 2*time.Minute + 3*time.Second,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := map[string]any{
				"id":       "test-id",
				"timeouts": []any{tt.timeouts},
			}
			d, err := NewResourceDataFromMaps(timeoutsTestSchema(), []string{"id"}, plan, nil, nil)
			require.NoError(t, err)

			got, err := getTimeoutValue(ctx, d, tt.timeoutType, fallbackTimeout)
			if tt.wantErrorSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrorSubstr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
