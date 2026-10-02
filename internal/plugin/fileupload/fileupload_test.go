package fileupload_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/fileupload"
)

func TestChecksum(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "data")
	require.NoError(t, os.WriteFile(path, []byte("abc"), 0o600))

	got, err := fileupload.Checksum(path)
	require.NoError(t, err)
	// sha256("abc")
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", got)

	_, err = fileupload.Checksum(filepath.Join(t.TempDir(), "missing"))
	require.ErrorContains(t, err, "failed to open file")
}

func TestUpload(t *testing.T) {
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

			path := filepath.Join(t.TempDir(), "payload")
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

			err := fileupload.Upload(t.Context(), path, checksum, contentType, server.URL)

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

		err := fileupload.Upload(t.Context(), t.TempDir(), checksum, contentType, "http://127.0.0.1:0")
		require.ErrorContains(t, err, "directory")
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		err := fileupload.Upload(t.Context(), filepath.Join(t.TempDir(), "nope"), checksum, contentType, "http://127.0.0.1:0")
		require.ErrorContains(t, err, "failed to open file")
	})

	t.Run("computes checksum when empty", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "payload")
		require.NoError(t, os.WriteFile(path, []byte("jar"), 0o600))

		var seenChecksum string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenChecksum = r.Header.Get("Content-SHA256")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		require.NoError(t, fileupload.Upload(t.Context(), path, "", contentType, server.URL))
		// sha256("jar")
		require.Equal(t, "0163f1eea7894350060624d315234d40c508ab251ba121714e234503045faadd", seenChecksum)
	})
}
