package valkeyuser

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
)

func TestCreateView(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	client.EXPECT().
		ServiceUserCreate(mock.Anything, "my-project", "my-valkey", &service.ServiceUserCreateIn{
			Username: "myuser",
			AccessControl: &service.AccessControlIn{
				ValkeyAclCategories: &[]string{"+@all"},
				ValkeyAclChannels:   &[]string{},
				ValkeyAclCommands:   &[]string{"+get"},
				ValkeyAclDatabases:  &[]string{},
				ValkeyAclKeys:       &[]string{"*"},
			},
		}).
		Return(&service.ServiceUserCreateOut{}, nil).
		Once()
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":               "my-project",
			"service_name":          "my-valkey",
			"username":              "myuser",
			"valkey_acl_categories": []any{"+@all"},
			"valkey_acl_commands":   []any{"+get"},
			"valkey_acl_keys":       []any{"*"},
		}),
	)
	require.NoError(t, err)
	require.NoError(t, ResourceOptions.Create(ctx, client, d))
}

func TestForceACLSlices(t *testing.T) {
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields)
	require.NoError(t, err)
	dto := map[string]any{"valkey_acl_keys": []any{"*"}}
	require.NoError(t, forceACLSlices(d, dto))
	require.Equal(t, map[string]any{
		"valkey_acl_categories": []any{},
		"valkey_acl_channels":   []any{},
		"valkey_acl_commands":   []any{},
		"valkey_acl_databases":  []any{},
		"valkey_acl_keys":       []any{"*"},
	}, dto)

	t.Run("skips the hoisted non-list fields", func(t *testing.T) {
		sch := resourceSchemaInternal()
		sch.Properties["valkey_acl_flag"] = &adapter.Schema{Type: adapter.SchemaTypeBool, HoistedFrom: serviceuser.AccessControlField}
		d, err := adapter.NewResourceData(sch, ResourceOptions.IdentityFields)
		require.NoError(t, err)
		dto := map[string]any{}
		require.NoError(t, forceACLSlices(d, dto))
		require.NotContains(t, dto, "valkey_acl_flag")
	})
}

func TestFlattenModifier(t *testing.T) {
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"project":         "my-project",
			"service_name":    "my-valkey",
			"username":        "myuser",
			"valkey_acl_keys": []any{"*"},
		}),
	)
	require.NoError(t, err)
	rsp := map[string]any{"username": "myuser", "access_control": map[string]any{}}
	require.NoError(t, d.Flatten(rsp, flattenModifier(context.Background(), nil)))
	assert.Equal(t, []any{}, d.Get("valkey_acl_keys"), "an ACL dropped in Aiven must drift")
}

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
			ResourceOptions.IdentityFields,
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

	t.Run("databases change", func(t *testing.T) {
		ctx := context.Background()
		client := avngen.NewMockClient(t)
		client.EXPECT().
			ServiceUserCredentialsModify(
				mock.Anything,
				"my-project",
				"my-valkey",
				"myuser",
				mock.MatchedBy(func(in *service.ServiceUserCredentialsModifyIn) bool {
					return in.AccessControl != nil &&
						in.AccessControl.ValkeyAclDatabases != nil &&
						assert.ObjectsAreEqual([]string{"0", "3"}, *in.AccessControl.ValkeyAclDatabases)
				}),
			).
			Return(&service.ServiceUserCredentialsModifyOut{}, nil).
			Once()
		plan := values([]any{"+@all"})
		plan["valkey_acl_databases"] = []any{"0", "3"}
		d := newResourceData(t, plan, values([]any{"+@all"}))
		require.NoError(t, ResourceOptions.Update(ctx, client, d))
	})
}
