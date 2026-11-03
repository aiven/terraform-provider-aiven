package clickhouseuser

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// An update without a password change would reset the password: the mock fails on any call.
func TestUpdateWithoutPasswordChange(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id":           "my-project/my-clickhouse/abc-uuid",
			"project":      "my-project",
			"service_name": "my-clickhouse",
			"uuid":         "abc-uuid",
			"password":     "secret",
		}
	}
	plan := values()
	plan["timeouts"] = map[string]any{"update": "20m"}
	state := values()
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(plan),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)
	err = ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
	require.ErrorIs(t, err, adapter.ErrNoChanges)
}
