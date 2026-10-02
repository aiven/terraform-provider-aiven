package acctest

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/noop-sink-connector.jar
var noopSinkConnectorJar []byte

// KafkaConnectPluginJar writes a minimal Kafka Connect sink connector jar to a temporary file
// and returns its path. The backend verifies plugin files by scanning for a ServiceLoader-
// discoverable connector class, so an archive-only jar won't do: this one bundles
// META-INF/services/org.apache.kafka.connect.sink.SinkConnector plus the matching .class file.
func KafkaConnectPluginJar(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "noop-sink-connector.jar")
	require.NoError(t, os.WriteFile(path, noopSinkConnectorJar, 0o600), "writing plugin jar failed")
	return path
}
