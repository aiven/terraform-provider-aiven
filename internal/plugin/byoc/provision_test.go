package byoc

import (
	"errors"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	byochandler "github.com/aiven/go-client-codegen/handler/byoc"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestRejectIdentityChange(t *testing.T) {
	t.Parallel()

	const (
		orgID    = "org1a23f456789"
		envID    = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
		oldEnvID = "4d5e6f7a-8b9c-4d0e-1f2a-3b4c5d6e7f8a"
		subID    = "2b3c4d5e-6f7a-4b8c-9d0e-1f2a3b4c5d6e"
		otherSub = "3c4d5e6f-7a8b-4c9d-0e1f-2a3b4c5d6e7f"
	)
	sch := &adapter.Schema{
		Type: adapter.SchemaTypeObject,
		Properties: map[string]*adapter.Schema{
			"organization_id":             {Type: adapter.SchemaTypeString},
			"custom_cloud_environment_id": {Type: adapter.SchemaTypeString},
			"azure_subscription_id":       {Type: adapter.SchemaTypeString},
		},
	}
	getErr := errors.New("get failed")

	cases := []struct {
		name    string
		envID   string // Empty is unknown: no API call
		planned string // Empty is unknown
		server  *string
		getErr  error
		wantErr bool
	}{
		{name: "entity created or replaced", planned: subID},
		{name: "not provisioned", envID: envID, planned: subID},
		{name: "same", envID: envID, planned: subID, server: lo.ToPtr(subID)},
		{name: "different", envID: envID, planned: subID, server: lo.ToPtr(otherSub), wantErr: true},
		{name: "unknown planned value", envID: envID, server: lo.ToPtr(subID)},
		{name: "get error", envID: envID, planned: subID, getErr: getErr, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Unknown values are tftypes.Value in the plan and missing from the config, as in ModifyPlan
			plan := map[string]any{"organization_id": orgID}
			config := map[string]any{"organization_id": orgID}
			for k, v := range map[string]string{"custom_cloud_environment_id": c.envID, "azure_subscription_id": c.planned} {
				if v == "" {
					plan[k] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
				} else {
					plan[k], config[k] = v, v
				}
			}
			d, err := adapter.NewResourceData(sch, []string{"organization_id", "custom_cloud_environment_id"},
				adapter.WithPreservePlanValues(),
				adapter.WithTestPlan(plan),
				adapter.WithTestConfig(config),
				// The environment before the entity was replaced: never read
				adapter.WithTestState(map[string]any{
					"organization_id":             orgID,
					"custom_cloud_environment_id": oldEnvID,
					"azure_subscription_id":       otherSub,
				}),
			)
			require.NoError(t, err)

			client := avngen.NewMockClient(t)
			if c.envID != "" {
				client.EXPECT().CustomCloudEnvironmentGet(t.Context(), orgID, c.envID).
					Return(&byochandler.CustomCloudEnvironmentGetOut{AzureSubscriptionId: c.server}, c.getErr)
			}

			err = RejectIdentityChange("azure_subscription_id")(t.Context(), client, d)
			if !c.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if c.getErr != nil {
				require.ErrorIs(t, err, c.getErr)
			} else {
				require.ErrorContains(t, err, "azure_subscription_id")
			}
		})
	}
}
