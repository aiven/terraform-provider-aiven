package acctest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CopyFile reads the file at source and writes a copy into the test's tempdir under name,
// returning the new path. Useful for giving a test its own mutable copy of a jar fixture.
func CopyFile(t *testing.T, source, name string) string {
	t.Helper()

	b, err := os.ReadFile(source) //nolint:gosec // test-controlled path
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, b, 0o600))
	return path
}

// AppendToFile appends content to the file at path, changing its sha256. Pairs with CopyFile
// to simulate an edited upload between test steps.
func AppendToFile(t *testing.T, path, content string) {
	t.Helper()

	file, err := os.OpenFile(filepath.Clean(path), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString(content)
	require.NoError(t, err)
}
