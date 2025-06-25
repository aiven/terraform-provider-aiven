package changelog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUpdate validates the changelog update functionality.
func TestUpdate(t *testing.T) {
	testCases := []struct {
		name            string
		initialContent  string
		version         string
		date            string
		expectedContent string
		expectErr       bool
		expectedErrType error
	}{
		{
			name: "successful update",
			initialContent: `# Changelog

All notable changes to this project will be documented in this file.

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

### Enhancements

- Added a new feature.
`,
			version: "1.2.3",
			date:    "2025-06-23",
			expectedContent: `# Changelog

All notable changes to this project will be documented in this file.

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

## [1.2.3] - 2025-06-23

### Enhancements

- Added a new feature.
`,
			expectErr: false,
		},
		{
			name:            "file does not exist",
			initialContent:  "", // no file will be created
			version:         "1.2.3",
			date:            "2025-06-23",
			expectErr:       true,
			expectedErrType: os.ErrNotExist,
		},
		{
			name:            "template header missing",
			initialContent:  "# Changelog\n\nSome content without the template header.\n",
			version:         "1.2.3",
			date:            "2025-06-23",
			expectErr:       true,
			expectedErrType: ErrTemplateHeaderNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			changelogPath := filepath.Join(tempDir, "CHANGELOG.md")

			if tc.initialContent != "" {
				err := os.WriteFile(changelogPath, []byte(tc.initialContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to write initial changelog file: %v", err)
				}
			}

			err := Update(changelogPath, tc.version, tc.date)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected an error, but got nil")
				}
				// Use errors.Is for robust checking against wrapped/sentinel errors
				if !errors.Is(err, tc.expectedErrType) {
					t.Errorf("Expected error of type %T, but got a different type: %v", tc.expectedErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Did not expect an error, but got: %v", err)
			}

			finalContent, readErr := os.ReadFile(changelogPath)
			if readErr != nil {
				t.Fatalf("Failed to read back changelog file: %v", readErr)
			}

			if string(finalContent) != tc.expectedContent {
				t.Errorf("File content mismatch.\nExpected:\n%s\n\nGot:\n%s", tc.expectedContent, string(finalContent))
			}
		})
	}
}

// TestExtract validates the changelog extraction functionality.
func TestExtract(t *testing.T) {
	const changelogContent = `# Changelog

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

## [2.0.0] - 2025-05-20

### BREAKING CHANGES

- Dropped support for old API.

### Features

- Added a major new feature.

## [1.1.0] - 2025-04-15

### Enhancements

- Improved performance of an existing feature.

## [1.0.1] - 2025-04-10

- Fixed a bug.

## [1.0.0] - 2025-04-01
`

	testCases := []struct {
		name             string
		versionToExtract string
		expectedOutput   string
		expectErr        bool
		errContains      string
	}{
		{
			name:             "extract top version with multiple sections",
			versionToExtract: "2.0.0",
			expectedOutput: `### BREAKING CHANGES

- Dropped support for old API.

### Features

- Added a major new feature.`,
		},
		{
			name:             "extract middle version",
			versionToExtract: "1.1.0",
			expectedOutput: `### Enhancements

- Improved performance of an existing feature.`,
		},
		{
			name:             "extract version with single line content",
			versionToExtract: "1.0.1",
			expectedOutput:   `- Fixed a bug.`,
		},
		{
			name:             "extract version with no content",
			versionToExtract: "1.0.0",
			expectErr:        true,
			errContains:      "no content found for version \"1.0.0\"",
		},
		{
			name:             "extract non-existent version",
			versionToExtract: "9.9.9",
			expectErr:        true,
			errContains:      "version \"9.9.9\" not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			changelogPath := filepath.Join(tempDir, "CHANGELOG.md")

			err := os.WriteFile(changelogPath, []byte(changelogContent), 0o644)
			if err != nil {
				t.Fatalf("Failed to write initial changelog file: %v", err)
			}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			extractErr := Extract(changelogPath, tc.versionToExtract)

			w.Close()
			os.Stdout = oldStdout

			var capturedOutput bytes.Buffer
			_, _ = io.Copy(&capturedOutput, r)

			if tc.expectErr {
				if extractErr == nil {
					t.Fatal("Expected an error, but got nil")
				}
				if !strings.Contains(extractErr.Error(), tc.errContains) {
					t.Errorf("Expected error to contain %q, but got %q", tc.errContains, extractErr.Error())
				}
				return
			}

			if extractErr != nil {
				t.Fatalf("Did not expect an error, but got: %v", extractErr)
			}

			if strings.TrimSpace(capturedOutput.String()) != strings.TrimSpace(tc.expectedOutput) {
				t.Errorf("Captured output mismatch.\nExpected:\n%s\n\nGot:\n%s", tc.expectedOutput, capturedOutput.String())
			}
		})
	}
}
