// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func testTimeoutsMap(kw ...string) map[string]attr.Value {
	base := map[string]attr.Value{
		"create":  types.StringNull(),
		"read":    types.StringNull(),
		"update":  types.StringNull(),
		"delete":  types.StringNull(),
		"default": types.StringNull(),
	}
	for i := 0; i < len(kw); i += 2 {
		if i+1 < len(kw) {
			base[kw[i]] = types.StringValue(kw[i+1])
		}
	}
	return base
}

func TestGetTimeoutValue(t *testing.T) {
	t.Parallel()

	const fallbackTimeout = 5 * time.Minute
	tests := []struct {
		name        string
		timeouts    map[string]attr.Value
		timeoutType timeoutType
		want        time.Duration
		wantError   bool
	}{
		{
			name:        "valid create timeout",
			timeouts:    testTimeoutsMap("create", "30s"),
			timeoutType: timeoutCreate,
			want:        30 * time.Second,
		},
		{
			name:        "valid read timeout",
			timeouts:    testTimeoutsMap("read", "45s"),
			timeoutType: timeoutRead,
			want:        45 * time.Second,
		},
		{
			name:        "valid update timeout",
			timeouts:    testTimeoutsMap("update", "1m"),
			timeoutType: timeoutUpdate,
			want:        1 * time.Minute,
		},
		{
			name:        "valid delete timeout",
			timeouts:    testTimeoutsMap("delete", "2m"),
			timeoutType: timeoutDelete,
			want:        2 * time.Minute,
		},
		{
			name:        "valid default timeout",
			timeouts:    testTimeoutsMap("default", "1m"),
			timeoutType: timeoutCreate,
			want:        1 * time.Minute,
		},
		{
			name:        "invalid duration format",
			timeouts:    testTimeoutsMap("create", "invalid"),
			timeoutType: timeoutCreate,
			want:        0,
			wantError:   true,
		},
		{
			name:        "uses default when timeout not specified",
			timeouts:    testTimeoutsMap("update", "45s"),
			timeoutType: timeoutCreate,
			want:        fallbackTimeout,
		},
		{
			name:        "empty timeouts uses fallback",
			timeouts:    testTimeoutsMap(),
			timeoutType: timeoutCreate,
			want:        fallbackTimeout,
		},
		{
			name:        "complex duration format",
			timeouts:    testTimeoutsMap("create", "1h2m3s"),
			timeoutType: timeoutCreate,
			want:        1*time.Hour + 2*time.Minute + 3*time.Second,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeoutsObj, diags := types.ObjectValue(
				map[string]attr.Type{
					"create":  types.StringType,
					"read":    types.StringType,
					"update":  types.StringType,
					"delete":  types.StringType,
					"default": types.StringType,
				},
				tt.timeouts,
			)
			require.Empty(t, diags)

			got, diags := getTimeoutValue(ctx, timeoutsObj, tt.timeoutType, fallbackTimeout)
			if tt.wantError {
				require.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			require.Equal(t, tt.want, got)
		})
	}
}
