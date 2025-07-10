package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		expected     string
		shouldChange bool
		fileMode     os.FileMode
	}{
		{
			name:         "remove trailing spaces",
			input:        "line with spaces   \nanother line  \n",
			expected:     "line with spaces\nanother line\n",
			shouldChange: true,
			fileMode:     0o644,
		},
		{
			name:         "remove trailing tabs",
			input:        "line with tabs\t\t\nanother line\t\n",
			expected:     "line with tabs\nanother line\n",
			shouldChange: true,
			fileMode:     0o644,
		},
		{
			name:         "mixed trailing whitespace",
			input:        "line with mixed \t \nanother line \t\n",
			expected:     "line with mixed\nanother line\n",
			shouldChange: true,
			fileMode:     0o644,
		},
		{
			name:         "preserve file ending with newline",
			input:        "line one\nline two\n",
			expected:     "line one\nline two\n",
			shouldChange: false,
			fileMode:     0o644,
		},
		{
			name:         "preserve file ending without newline",
			input:        "line one\nline two",
			expected:     "line one\nline two",
			shouldChange: false,
			fileMode:     0o644,
		},
		{
			name:         "file with only trailing whitespace",
			input:        "clean line\nline with spaces   ",
			expected:     "clean line\nline with spaces",
			shouldChange: true,
			fileMode:     0o644,
		},
		{
			name:         "empty file",
			input:        "",
			expected:     "",
			shouldChange: false,
			fileMode:     0o644,
		},
		{
			name:         "single line with trailing space",
			input:        "single line   ",
			expected:     "single line",
			shouldChange: true,
			fileMode:     0o644,
		},
		{
			name:         "executable file permissions",
			input:        "#!/bin/bash\necho hello   \n",
			expected:     "#!/bin/bash\necho hello\n",
			shouldChange: true,
			fileMode:     0o755,
		},
		{
			name:         "preserve multiple trailing newlines behavior",
			input:        "line one\nline two\n\n",
			expected:     "line one\nline two\n",
			shouldChange: true,
			fileMode:     0o644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.txt")

			err := os.WriteFile(tmpFile, []byte(tt.input), tt.fileMode)
			require.NoError(t, err, "Failed to create temp file")

			originalInfo, err := os.Stat(tmpFile)
			require.NoError(t, err, "Failed to stat temp file")

			err = formatFile(tmpFile)
			require.NoError(t, err)

			result, err := os.ReadFile(tmpFile)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(result), "Formatted content does not match expected output")

			newInfo, err := os.Stat(tmpFile)
			require.NoError(t, err, "Failed to stat file after formatting")

			assert.Equal(t, originalInfo.Mode(), newInfo.Mode(), "File permissions should be preserved")
		})
	}
}

func TestFormatFileErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
	}{
		{
			name: "nonexistent file",
			setupFunc: func(_ *testing.T) string {
				return "/nonexistent/path/file.txt"
			},
			expectError: true,
		},
		{
			name: "directory instead of file",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setupFunc(t)
			err := formatFile(path)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	t.Parallel()

	expectedExtensions := []string{".go", ".yml", ".yaml", ".md", ".json"}
	expectedExcludeFiles := []string{"definitions/.schema.yml"}
	expectedExcludeDirs := []string{".git", "vendor"}

	assert.Equal(t, expectedExtensions, defaultExtensions, "Default extensions should match expected values")
	assert.Equal(t, expectedExcludeFiles, defaultExcludeFiles, "Default exclude files should match expected values")
	assert.Equal(t, expectedExcludeDirs, defaultExcludeDirs, "Default exclude directories should match expected values")
}

func TestProcessFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFiles := map[string]string{
		"file1.txt": "line with spaces   \nclean line\n",
		"file2.txt": "another line  \nclean line\n",
		"file3.txt": "clean file\nno changes needed\n",
	}

	filePaths := make([]string, 0, len(testFiles))
	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err, "Failed to create test file %s", filename)
		filePaths = append(filePaths, path)
	}

	originalWorkers := workers
	workers = 2
	defer func() { workers = originalWorkers }()

	err := processFiles(filePaths)
	require.NoError(t, err, "processFiles should not fail")

	expectedResults := map[string]string{
		"file1.txt": "line with spaces\nclean line\n",
		"file2.txt": "another line\nclean line\n",
		"file3.txt": "clean file\nno changes needed\n",
	}

	for filename, expected := range expectedResults {
		path := filepath.Join(tmpDir, filename)
		result, err := os.ReadFile(path)
		require.NoError(t, err, "Failed to read result file %s", filename)

		assert.Equal(t, expected, string(result), "File %s content should match expected output", filename)
	}
}

func TestProcessFilesWithError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	problemFile := filepath.Join(tmpDir, "problem.txt")

	err := os.WriteFile(problemFile, []byte("content"), 0o644)
	require.NoError(t, err, "Failed to create problem file")

	// make file unreadable
	err = os.Chmod(problemFile, 0o000)
	require.NoError(t, err, "Failed to change file permissions")

	// restore permissions after test
	defer func(name string, mode os.FileMode) {
		_ = os.Chmod(name, mode)
	}(problemFile, 0o644)

	originalWorkers := workers
	workers = 1
	defer func() { workers = originalWorkers }()

	err = processFiles([]string{problemFile})
	assert.Error(t, err, "Expected an error when processing unreadable file")
}

func TestRootDirFlag(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err, "Failed to create subdirectory")

	rootFile := filepath.Join(tmpDir, "root.go")
	subFile := filepath.Join(subDir, "sub.go")

	rootContent := "package main   \n"
	subContent := "package sub  \n"

	err = os.WriteFile(rootFile, []byte(rootContent), 0o644)
	require.NoError(t, err, "Failed to create root file")

	err = os.WriteFile(subFile, []byte(subContent), 0o644)
	require.NoError(t, err, "Failed to create sub file")

	// test with different root directories
	tests := []struct {
		name           string
		rootDir        string
		expectedFiles  int
		shouldFindRoot bool
		shouldFindSub  bool
	}{
		{
			name:           "root directory processes all files",
			rootDir:        tmpDir,
			expectedFiles:  2,
			shouldFindRoot: true,
			shouldFindSub:  true,
		},
		{
			name:           "subdirectory processes only subdirectory files",
			rootDir:        subDir,
			expectedFiles:  1,
			shouldFindRoot: false,
			shouldFindSub:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset file contents before each test
			_ = os.WriteFile(rootFile, []byte(rootContent), 0o644)
			_ = os.WriteFile(subFile, []byte(subContent), 0o644)

			originalRootDir := rootDir
			originalWorkers := workers

			rootDir = tt.rootDir
			workers = 2

			defer func() {
				rootDir = originalRootDir
				workers = originalWorkers
			}()

			extensions := defaultExtensions
			var filesToFormat []string
			err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				for _, excludeDir := range defaultExcludeDirs {
					if strings.Contains(path, excludeDir) {
						return nil
					}
				}

				for _, excludeFile := range defaultExcludeFiles {
					if strings.Contains(path, excludeFile) {
						return nil
					}
				}

				if !info.IsDir() {
					ext := filepath.Ext(path)
					for _, targetExt := range extensions {
						if ext == targetExt {
							filesToFormat = append(filesToFormat, path)
							break
						}
					}
				}

				return nil
			})

			require.NoError(t, err, "Failed to walk directory")

			assert.Len(t, filesToFormat, tt.expectedFiles, "Number of files to format should match expected")

			err = processFiles(filesToFormat)
			require.NoError(t, err, "Failed to process files")

			if tt.shouldFindRoot {
				content, err := os.ReadFile(rootFile)
				require.NoError(t, err, "Failed to read root file")
				expected := "package main\n"
				assert.Equal(t, expected, string(content), "Root file should be formatted correctly")
			}

			if tt.shouldFindSub {
				content, err := os.ReadFile(subFile)
				require.NoError(t, err, "Failed to read sub file")
				expected := "package sub\n"
				assert.Equal(t, expected, string(content), "Sub file should be formatted correctly")
			}
		})
	}
}
