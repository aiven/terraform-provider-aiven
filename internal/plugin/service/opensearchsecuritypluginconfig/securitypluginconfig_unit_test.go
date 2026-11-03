package opensearchsecuritypluginconfig

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/opensearch"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestUpdateView(t *testing.T) {
	values := func(password string) map[string]any {
		return map[string]any{
			"id":             "my-project/my-os",
			"project":        "my-project",
			"service_name":   "my-os",
			"admin_password": password,
		}
	}
	newResourceData := func(t *testing.T, plan, state map[string]any) adapter.ResourceData {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestPlan(plan),
			adapter.WithTestState(state),
		)
		require.NoError(t, err)
		return d
	}

	// The mock fails on any call
	t.Run("no changes", func(t *testing.T) {
		plan := values("secret")
		plan["timeouts"] = map[string]any{"update": "20m"}
		d := newResourceData(t, plan, values("secret"))
		err := ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
		require.ErrorIs(t, err, adapter.ErrNoChanges)
	})

	t.Run("password change", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		client.EXPECT().
			ServiceOpenSearchSecurityReset(mock.Anything, "my-project", "my-os", &opensearch.ServiceOpenSearchSecurityResetIn{
				AdminPassword: "old",
				NewPassword:   "new",
			}).
			Return(&opensearch.ServiceOpenSearchSecurityResetOut{}, nil).
			Once()
		d := newResourceData(t, values("new"), values("old"))
		require.NoError(t, ResourceOptions.Update(context.Background(), client, d))
	})
}
