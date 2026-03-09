package adapter

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestResourceAdapter_refreshState(t *testing.T) {
	warning := func(summary string) diag.Diagnostic {
		return diag.NewWarningDiagnostic(summary, "detail")
	}

	t.Run("keeps outer warnings and merges only last try warnings", func(t *testing.T) {
		var target diag.Diagnostics

		outerCtx, drainWarnings := withWarnings(t.Context(), &target)
		AddWarnings(outerCtx, warning("outer-1"), warning("outer-2"))

		outerCollector, ok := outerCtx.Value(warningCollectorKey{}).(*warningCollector)
		require.True(t, ok)
		require.NotNil(t, outerCollector)

		warningsByTry := [][]diag.Diagnostic{
			{warning("try-1-a"), warning("try-1-b")},
			{warning("try-2-a"), warning("try-2-b")},
		}

		attempts := 0
		a := &resourceAdapter{
			resource: ResourceOptions{
				Read: func(ctx context.Context, _ avngen.Client, _ ResourceData) error {
					attempts++
					AddWarnings(ctx, warningsByTry[attempts-1]...)

					if attempts == 1 {
						return avngen.Error{Status: http.StatusNotFound, Message: "not ready"}
					}

					return nil
				},
			},
		}

		err := a.refreshState(outerCtx, nil)
		drainWarnings()

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
		require.Equal(t, 4, target.WarningsCount())
		require.True(t, target.Contains(warning("outer-1")))
		require.True(t, target.Contains(warning("outer-2")))
		require.True(t, target.Contains(warning("try-2-a")))
		require.True(t, target.Contains(warning("try-2-b")))
		require.False(t, target.Contains(warning("try-1-a")))
		require.False(t, target.Contains(warning("try-1-b")))
	})
}
