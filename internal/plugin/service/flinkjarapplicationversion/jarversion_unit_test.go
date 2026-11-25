package flinkjarapplicationversion

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestModifyPlanUnknownSource(t *testing.T) {
	t.Parallel()

	t.Run("new resource does not require replacement", func(t *testing.T) {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			ResourceOptions.IdentityFields,
			adapter.WithTestPlan(map[string]any{
				"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		)
		require.NoError(t, err)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Empty(t, adapter.RequiresReplaceKeys(d))
	})

	t.Run("existing resource requires replacement", func(t *testing.T) {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			ResourceOptions.IdentityFields,
			adapter.WithTestPlan(map[string]any{
				"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			adapter.WithTestState(map[string]any{
				"id":            "project/service/application/version",
				fileSha256Field: "old-checksum",
			}),
		)
		require.NoError(t, err)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{"source", fileSha256Field}, adapter.RequiresReplaceKeys(d))
	})
}
