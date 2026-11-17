package byoc

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestDeleteProvisioned(t *testing.T) {
	t.Parallel()

	const orgID, envID = "org1a23f456789", "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
	sch := &adapter.Schema{
		Type: adapter.SchemaTypeObject,
		Properties: map[string]*adapter.Schema{
			"organization_id":             {Type: adapter.SchemaTypeString},
			"custom_cloud_environment_id": {Type: adapter.SchemaTypeString},
		},
	}

	cases := []struct {
		state   byoc.CustomCloudEnvironmentStateType
		deletes int
	}{
		{byoc.CustomCloudEnvironmentStateTypeActive, 1},
		{byoc.CustomCloudEnvironmentStateTypeDisconnected, 1},
		{byoc.CustomCloudEnvironmentStateTypeCreationFailed, 0},
		{byoc.CustomCloudEnvironmentStateTypeCreating, 0},
		{byoc.CustomCloudEnvironmentStateTypeDraft, 0},
	}
	for _, c := range cases {
		t.Run(string(c.state), func(t *testing.T) {
			t.Parallel()

			d, err := adapter.NewResourceData(sch, nil, adapter.WithTestState(map[string]any{
				"organization_id":             orgID,
				"custom_cloud_environment_id": envID,
			}))
			require.NoError(t, err)

			client := avngen.NewMockClient(t)
			client.EXPECT().CustomCloudEnvironmentGet(t.Context(), orgID, envID).
				Return(&byoc.CustomCloudEnvironmentGetOut{State: c.state}, nil)

			deletes := 0
			del := DeleteProvisioned(func(context.Context, avngen.Client, adapter.ResourceData) error {
				deletes++
				return nil
			})

			err = del(t.Context(), client, d)
			require.Equal(t, c.deletes, deletes)
			if c.deletes == 0 {
				require.ErrorIs(t, err, adapter.ErrDeleteSkipped)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
