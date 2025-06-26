package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTest(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		path         string
		expectedName string
		expectedType string
	}{
		{
			name:         "standard plugin path",
			path:         filepath.Join(".", "internal", "plugin", "datasource", "kafka"),
			expectedName: "kafka",
			expectedType: "plugin",
		},
		{
			name:         "standard sdkprovider path",
			path:         filepath.Join(".", "internal", "sdkprovider", "service", "vpc"),
			expectedName: "vpc",
			expectedType: "sdk",
		},
		{
			name:         "generic path without prefix",
			path:         filepath.Join(".", "internal", "schema", "util"),
			expectedName: "util",
			expectedType: "generic",
		},
		{
			name:         "deeply nested sdkprovider path",
			path:         filepath.Join(".", "internal", "sdkprovider", "very", "deep", "path"),
			expectedName: "path",
			expectedType: "sdk",
		},
		{
			name:         "root internal path component",
			path:         filepath.Join(".", "internal", "client"),
			expectedName: "client",
			expectedType: "generic",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test := newTest(tc.path)
			assert.Equal(t, tc.path, test.Path)
			assert.Equal(t, tc.expectedName, test.Name)
			assert.Equal(t, tc.expectedType, test.Type)
		})
	}
}

func TestPathToUniqueName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		root     string
		path     string
		expected string
	}{
		{
			name:     "plugin path",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "internal", "plugin", "datasource", "foo"),
			expected: "plugin-datasource-foo",
		},
		{
			name:     "sdkprovider path with name collision potential",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "internal", "sdkprovider", "service", "common"),
			expected: "sdk-service-common",
		},
		{
			name:     "plugin path with name collision potential",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "internal", "plugin", "service", "common"),
			expected: "plugin-service-common",
		},
		{
			name:     "generic path",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "internal", "acme", "client"),
			expected: "acme-client",
		},
		{
			name:     "path with single component under internal",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "internal", "foo"),
			expected: "foo",
		},
		{
			name:     "root from different base directory - plugin",
			root:     filepath.Join(".", "src"),
			path:     filepath.Join(".", "src", "plugin", "service", "test"),
			expected: "plugin-service-test",
		},
		{
			name:     "root from different base directory - sdk",
			root:     filepath.Join(".", "lib"),
			path:     filepath.Join(".", "lib", "sdkprovider", "client"),
			expected: "sdk-client",
		},
		{
			name:     "path outside root uses relative path",
			root:     filepath.Join(".", "internal"),
			path:     filepath.Join(".", "outside", "test"),
			expected: "..-outside-test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := pathToUniqueName(tc.root, tc.path)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGenerateMatrix(t *testing.T) {
	t.Parallel()

	tempRoot := t.TempDir()
	internalDir := filepath.Join(tempRoot, "internal")

	testDirs := []string{
		"plugin/service/common",
		"sdkprovider/service/common",
		"sdkprovider/service/kafka",
		"acme/client",
		"plugin/deep/nested/test",
	}

	for _, dir := range testDirs {
		fullDir := filepath.Join(internalDir, dir)
		require.NoError(t, os.MkdirAll(fullDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(fullDir, "dummy_test.go"), []byte("package dummy"), 0o644))
	}

	// test multiple files in one directory
	multiFileDir := filepath.Join(internalDir, "acme", "client")
	require.NoError(t, os.WriteFile(filepath.Join(multiFileDir, "extra_test.go"), []byte("package dummy"), 0o644))

	// directory with no test files (should be ignored)
	require.NoError(t, os.MkdirAll(filepath.Join(internalDir, "no-tests-here"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(internalDir, "no-tests-here", "code.go"), []byte("package dummy"), 0o644))

	allTests := []string{"plugin-service-common", "sdk-service-common", "sdk-service-kafka", "acme-client", "plugin-deep-nested-test"}

	testCases := []struct {
		name         string
		slowTestsCSV string
		expectedSlow []string
	}{
		{
			name:         "no slow tests",
			slowTestsCSV: "",
			expectedSlow: []string{},
		},
		{
			name:         "one slow test",
			slowTestsCSV: "sdk-service-kafka",
			expectedSlow: []string{"sdk-service-kafka"},
		},
		{
			name:         "slow test with ambiguity (sdk)",
			slowTestsCSV: "sdk-service-common",
			expectedSlow: []string{"sdk-service-common"},
		},
		{
			name:         "slow test with ambiguity (plugin)",
			slowTestsCSV: "plugin-service-common",
			expectedSlow: []string{"plugin-service-common"},
		},
		{
			name:         "multiple slow tests with whitespace",
			slowTestsCSV: "  acme-client, plugin-deep-nested-test  ",
			expectedSlow: []string{"acme-client", "plugin-deep-nested-test"},
		},
		{
			name:         "all tests slow",
			slowTestsCSV: strings.Join(allTests, ","),
			expectedSlow: allTests,
		},
		{
			name:         "non-existent slow test",
			slowTestsCSV: "this-does-not-exist",
			expectedSlow: []string{},
		},
		{
			name:         "empty csv string",
			slowTestsCSV: ", ,",
			expectedSlow: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			matrix, err := GenerateMatrix(internalDir, tc.slowTestsCSV)
			require.NoError(t, err)

			getUniqueNames := func(tests []Test) []string {
				names := make([]string, len(tests))
				for i, test := range tests {
					names[i] = pathToUniqueName(internalDir, test.Path)
				}
				return names
			}

			slowSet := make(map[string]struct{})
			for _, slow := range tc.expectedSlow {
				slowSet[slow] = struct{}{}
			}

			var expectedNormal []string
			for _, test := range allTests {
				if _, isSlow := slowSet[test]; !isSlow {
					expectedNormal = append(expectedNormal, test)
				}
			}

			assert.ElementsMatch(t, tc.expectedSlow, getUniqueNames(matrix.Slow), "mismatch in slow tests")
			assert.ElementsMatch(t, expectedNormal, getUniqueNames(matrix.Normal), "mismatch in normal tests")
		})
	}
}
