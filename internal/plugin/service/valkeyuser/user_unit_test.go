package valkeyuser

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestUpdateView(t *testing.T) {
	values := func(categories []any) map[string]any {
		return map[string]any{
			"id":                    "my-project/my-valkey/myuser",
			"project":               "my-project",
			"service_name":          "my-valkey",
			"username":              "myuser",
			"password":              "secret",
			"valkey_acl_categories": categories,
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
		plan := values([]any{"+@all"})
		plan["timeouts"] = map[string]any{"update": "20m"}
		d := newResourceData(t, plan, values([]any{"+@all"}))
		err := ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
		require.ErrorIs(t, err, adapter.ErrNoChanges)
	})

	t.Run("access control change", func(t *testing.T) {
		ctx := context.Background()
		client := avngen.NewMockClient(t)
		client.EXPECT().
			ServiceUserCredentialsModify(
				mock.Anything,
				"my-project",
				"my-valkey",
				"myuser",
				mock.MatchedBy(func(in *service.ServiceUserCredentialsModifyIn) bool {
					return in.Operation == service.ServiceUserCredentialsModifyOperationTypeSetAccessControl
				}),
			).
			Return(&service.ServiceUserCredentialsModifyOut{}, nil).
			Once()
		d := newResourceData(t, values([]any{"-@all"}), values([]any{"+@all"}))
		require.NoError(t, ResourceOptions.Update(ctx, client, d))
	})
}
