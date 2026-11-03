package clickhouseuser

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/clickhouse"
	"github.com/stretchr/testify/mock"
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
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(plan),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)
	err = ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
	require.ErrorIs(t, err, adapter.ErrNoChanges)
}

// Removing the write-only password sends the reset without a password: the API generates one.
func TestUpdateToGeneratedPassword(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id":           "my-project/my-clickhouse/abc-uuid",
			"project":      "my-project",
			"service_name": "my-clickhouse",
			"uuid":         "abc-uuid",
		}
	}
	state := values()
	state["password_wo_version"] = 2
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(values()),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)

	client := avngen.NewMockClient(t)
	client.EXPECT().
		ServiceClickHousePasswordReset(mock.Anything, "my-project", "my-clickhouse", "abc-uuid", &clickhouse.ServiceClickHousePasswordResetIn{}).
		Return("generated", nil).
		Once()
	require.NoError(t, ResourceOptions.Update(context.Background(), client, d))
}
