package jarversion

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
				"id":                "project/service/application/version",
				sourceChecksumField: "old-checksum",
			}),
		)
		require.NoError(t, err)

		tracked := &replacementTrackingResourceData{ResourceData: d}
		require.NoError(t, modifyPlan(t.Context(), nil, tracked))
		require.Equal(t, []string{sourceChecksumField}, tracked.requiresReplace)
	})
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	const checksum = "0f4e2b1a"

	cases := map[string]struct {
		status      int
		response    string
		expectedErr string
	}{
		"accepted":            {status: http.StatusOK},
		"rejected with body":  {status: http.StatusForbidden, response: "<Error>AccessDenied</Error>", expectedErr: `s3 error: 403 Forbidden: "<Error>AccessDenied</Error>"`},
		"rejected empty body": {status: http.StatusBadGateway, expectedErr: `s3 error: 502 Bad Gateway: ""`},
		"accepted with body":  {status: http.StatusOK, response: "unexpected", expectedErr: "s3 error: unexpected"},
	}

	type request struct {
		method   string
		checksum string
	}

	for name, opt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "app.jar")
			require.NoError(t, os.WriteFile(path, []byte("jar"), 0o600))

			// The channel hands the request to the assertions below: a test failure must not be
			// reported from the handler's goroutine.
			requests := make(chan request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- request{method: r.Method, checksum: r.Header.Get("Content-SHA256")}

				w.WriteHeader(opt.status)
				_, _ = w.Write([]byte(opt.response))
			}))
			defer server.Close()

			err := uploadFile(t.Context(), path, checksum, server.URL)

			got := <-requests
			require.Equal(t, http.MethodPut, got.method)
			require.Equal(t, checksum, got.checksum)

			if opt.expectedErr == "" {
				require.NoError(t, err)
				return
			}

			require.EqualError(t, err, opt.expectedErr)
		})
	}
}
