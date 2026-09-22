package custompluginfile

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

		want, err := fileChecksum(path)
		require.NoError(t, err)
		require.Equal(t, want, d.Get(sourceChecksumField))
	})

	t.Run("same content does not require replacement", func(t *testing.T) {
		t.Parallel()

		path := write(t)
		checksum, err := fileChecksum(path)
		require.NoError(t, err)

		d := newData(t,
			map[string]any{"source": path},
			map[string]any{"id": "org/file", sourceChecksumField: checksum},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Empty(t, d.requiresReplace)
	})

	t.Run("changed content requires replacement", func(t *testing.T) {
		t.Parallel()

		d := newData(t,
			map[string]any{"source": write(t)},
			map[string]any{"id": "org/file", sourceChecksumField: "old-checksum"},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{sourceChecksumField}, d.requiresReplace)
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
			map[string]any{"id": "org/file", sourceChecksumField: "old-checksum"},
		)

		require.NoError(t, modifyPlan(t.Context(), nil, d))
		require.Equal(t, []string{sourceChecksumField}, d.requiresReplace)
	})

	t.Run("missing file is an error", func(t *testing.T) {
		t.Parallel()

		d := newData(t, map[string]any{"source": filepath.Join(t.TempDir(), "missing.jar")}, nil)

		require.Error(t, modifyPlan(t.Context(), nil, d))
	})
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	const (
		checksum    = "0f4e2b1a"
		contentType = "application/zip"
	)

	cases := map[string]struct {
		status      int
		response    string
		expectedErr string
	}{
		"accepted":            {status: http.StatusOK},
		"rejected with body":  {status: http.StatusForbidden, response: "<Error>AccessDenied</Error>", expectedErr: `s3 upload error: 403 Forbidden: "<Error>AccessDenied</Error>"`},
		"rejected empty body": {status: http.StatusBadGateway, expectedErr: `s3 upload error: 502 Bad Gateway: ""`},
		"accepted with body":  {status: http.StatusOK, response: "unexpected", expectedErr: "s3 upload error: unexpected"},
	}

	type request struct {
		method      string
		checksum    string
		contentType string
		body        []byte
	}

	for name, opt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "plugin.zip")
			require.NoError(t, os.WriteFile(path, []byte("jar"), 0o600))

			// The channel hands the request to the assertions below: a test failure must not be
			// reported from the handler's goroutine.
			requests := make(chan request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := make([]byte, r.ContentLength)
				_, _ = r.Body.Read(body)
				requests <- request{
					method:      r.Method,
					checksum:    r.Header.Get("Content-SHA256"),
					contentType: r.Header.Get("Content-Type"),
					body:        body,
				}

				w.WriteHeader(opt.status)
				_, _ = w.Write([]byte(opt.response))
			}))
			defer server.Close()

			err := uploadFile(t.Context(), path, checksum, contentType, server.URL)

			got := <-requests
			require.Equal(t, http.MethodPut, got.method)
			require.Equal(t, checksum, got.checksum)
			require.Equal(t, contentType, got.contentType)
			require.Equal(t, []byte("jar"), got.body)

			if opt.expectedErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, opt.expectedErr)
		})
	}

	t.Run("directory", func(t *testing.T) {
		t.Parallel()

		err := uploadFile(t.Context(), t.TempDir(), checksum, contentType, "http://127.0.0.1:0")
		require.ErrorContains(t, err, "directory")
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		err := uploadFile(t.Context(), filepath.Join(t.TempDir(), "nope"), checksum, contentType, "http://127.0.0.1:0")
		require.ErrorContains(t, err, "failed to open file")
	})
}
