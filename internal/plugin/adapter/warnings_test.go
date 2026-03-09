package adapter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/require"
)

func TestWithWarnings(t *testing.T) {
	t.Parallel()

	t.Run("nil target is no-op", func(t *testing.T) {
		ctx, drainWarnings := withWarnings(t.Context(), nil)

		require.NotPanics(t, func() {
			AddWarning(ctx, "summary", "detail")
			AddAttributeWarning(ctx, path.Root("name"), "attr summary", "attr detail")
			AddWarnings(ctx, diag.NewWarningDiagnostic("summary", "detail"))
			drainWarnings()
		})
	})

	t.Run("drains collected warnings", func(t *testing.T) {
		var target diag.Diagnostics

		ctx, drainWarnings := withWarnings(t.Context(), &target)
		AddWarning(ctx, "summary", "detail")
		AddAttributeWarning(ctx, path.Root("name"), "attr summary", "attr detail")
		AddWarnings(ctx, diag.NewWarningDiagnostic("raw summary", "detail"))

		drainWarnings()

		require.Equal(t, 3, target.WarningsCount())
		require.True(t, target.Contains(diag.NewWarningDiagnostic("summary", "detail")))
		require.True(t, target.Contains(diag.NewAttributeWarningDiagnostic(
			path.Root("name"),
			"attr summary",
			"attr detail",
		)))
		require.True(t, target.Contains(diag.NewWarningDiagnostic("raw summary", "detail")))
	})
}
