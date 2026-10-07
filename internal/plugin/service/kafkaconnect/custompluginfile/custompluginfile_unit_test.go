package custompluginfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/fileupload"
)

type replacementTrackingResourceData struct {
	adapter.ResourceData
	requiresReplace []string
}

func (d *replacementTrackingResourceData) RequiresReplace(keys ...string) {
	d.ResourceData.RequiresReplace(keys...)
	d.requiresReplace = append(d.requiresReplace, keys...)
}

func TestModifyPlan(t *testing.T) {
	t.Parallel()

	const content = "jar"

	write := func(t *testing.T) string {
		t.Helper()

		path := filepath.Join(t.TempDir(), "plugin.jar")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}

	newData := func(t *testing.T, plan, state map[string]any) *replacementTrackingResourceData {
		t.Helper()

		opts := []adapter.ResourceDataOpt{adapter.WithTestPlan(plan)}
		if state != nil {
			opts = append(opts, adapter.WithTestState(state))
		}

		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), opts...)
		require.NoError(t, err)
		return &replacementTrackingResourceData{ResourceData: d}
	}

	t.Run("new resource stores the checksum without replacement", func(t *testing.T) {
		t.Parallel()

		path := write(t)
		d := newData(t, map[string]any{"source": path}, nil)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Empty(t, d.requiresReplace)

		want, err := fileupload.Checksum(path)
		require.NoError(t, err)
		require.Equal(t, want, d.Get(fileSha256Field))
	})

	t.Run("same content does not require replacement", func(t *testing.T) {
		t.Parallel()

		path := write(t)
		checksum, err := fileupload.Checksum(path)
		require.NoError(t, err)

		d := newData(t,
			map[string]any{"source": path},
			map[string]any{"id": "org/file", fileSha256Field: checksum},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Empty(t, d.requiresReplace)
	})

	t.Run("changed content requires replacement", func(t *testing.T) {
		t.Parallel()

		d := newData(t,
			map[string]any{"source": write(t)},
			map[string]any{"id": "org/file", fileSha256Field: "old-checksum"},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{"source", fileSha256Field}, d.requiresReplace)
	})

	t.Run("existing resource with empty prior checksum requires replacement", func(t *testing.T) {
		// Covers the backward-compat upgrade with -refresh=false: Read never runs to mirror the
		// backend hash, so prev stays empty. Without triggering replace, a local edit would
		// never be re-uploaded (createView is the only upload path).
		t.Parallel()

		d := newData(t,
			map[string]any{"source": write(t)},
			map[string]any{"id": "org/file"},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{"source", fileSha256Field}, d.requiresReplace)
	})

	t.Run("unknown source of a new resource does not require replacement", func(t *testing.T) {
		t.Parallel()

		d := newData(t, map[string]any{"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}, nil)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Empty(t, d.requiresReplace)
	})

	t.Run("unknown source of an existing resource requires replacement", func(t *testing.T) {
		t.Parallel()

		d := newData(t,
			map[string]any{"source": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
			map[string]any{"id": "org/file", fileSha256Field: "old-checksum"},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{"source", fileSha256Field}, d.requiresReplace)
	})

	t.Run("missing file is an error", func(t *testing.T) {
		t.Parallel()

		d := newData(t, map[string]any{"source": filepath.Join(t.TempDir(), "missing.jar")}, nil)

		require.Error(t, modifyPlan(t.Context(), nil, d))
	})
}
