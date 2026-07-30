package acctest

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// FlinkJarFile returns a jar file path to upload as a Flink jar application version.
// AIVEN_TEST_FLINK_JAR_FILE points to a real, runnable jar. Without it, a minimal jar passes
// the backend's archive check, but has no job to run.
func FlinkJarFile(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("AIVEN_TEST_FLINK_JAR_FILE"); path != "" {
		return path
	}

	path := filepath.Join(t.TempDir(), "minimal.jar")
	file, err := os.Create(path)
	require.NoError(t, err, "creating jar file failed")
	defer file.Close()

	zw := zip.NewWriter(file)

	// Directory entries must not be compressed.
	dir := &zip.FileHeader{Name: "META-INF/", Method: zip.Store, Modified: time.Now()}
	dir.SetMode(0o755)
	_, err = zw.CreateHeader(dir)
	require.NoError(t, err, "creating META-INF directory failed")

	manifest := &zip.FileHeader{Name: "META-INF/MANIFEST.MF", Method: zip.Deflate, Modified: time.Now()}
	manifest.SetMode(0o644)
	w, err := zw.CreateHeader(manifest)
	require.NoError(t, err, "creating manifest failed")

	_, err = io.WriteString(w, "Manifest-Version: 1.0\r\nCreated-By: test-jar-go\r\n\r\n")
	require.NoError(t, err, "writing manifest failed")
	require.NoError(t, zw.Close(), "closing jar file failed")

	return path
}
