package jarversion

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

type replacementTrackingResourceData struct {
	adapter.ResourceData
	requiresReplace []string
}

func (d *replacementTrackingResourceData) RequiresReplace(keys ...string) {
	d.ResourceData.RequiresReplace(keys...)
	d.requiresReplace = append(d.requiresReplace, keys...)
}

func TestModifyPlanUnknownSource(t *testing.T) {
	t.Parallel()

	t.Run("new resource does not require replacement", func(t *testing.T) {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestPlan(map[string]any{
				"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		)
		require.NoError(t, err)

		tracked := &replacementTrackingResourceData{ResourceData: d}
		require.NoError(t, modifyPlan(t.Context(), nil, tracked))
		require.Empty(t, tracked.requiresReplace)
	})

	t.Run("existing resource requires replacement", func(t *testing.T) {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestPlan(map[string]any{
				"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			adapter.WithTestState(map[string]any{
				"id":            "project/service/application/version",
				fileSha256Field: "old-checksum",
			}),
		)
		require.NoError(t, err)

		tracked := &replacementTrackingResourceData{ResourceData: d}
		require.NoError(t, modifyPlan(t.Context(), nil, tracked))
		require.Equal(t, []string{"source", fileSha256Field}, tracked.requiresReplace)
	})
}
