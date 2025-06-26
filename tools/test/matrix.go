package test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Test represents a single test suite found in a directory
type Test struct {
	Path string `json:"path"`
	Name string `json:"name"` // simple name, e.g., "kafka"
	Type string `json:"type"` // "plugin", "sdk", "generic"
}

// Matrix holds the partitioned lists of normal and slow tests
type Matrix struct {
	Normal []Test `json:"normal"`
	Slow   []Test `json:"slow"`
}

// GenerateMatrix discovers test suites, partitions them, and returns the matrix
func GenerateMatrix(root string, slowTestsCSV string) (*Matrix, error) {
	// slow tests lookup map
	slowTestsLookup := make(map[string]struct{})
	for _, testName := range strings.Split(slowTestsCSV, ",") {
		if trimmed := strings.TrimSpace(testName); trimmed != "" {
			slowTestsLookup[trimmed] = struct{}{}
		}
	}

	// find directories containing test files
	testDirs := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			testDirs[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directories: %w", err)
	}

	// partition tests into normal and slow
	matrix := &Matrix{}
	for dir := range testDirs {
		test := newTest(dir)
		uniqueName := pathToUniqueName(root, dir)

		if _, isSlow := slowTestsLookup[uniqueName]; isSlow {
			matrix.Slow = append(matrix.Slow, test)
		} else {
			matrix.Normal = append(matrix.Normal, test)
		}
	}

	return matrix, nil
}

// ToJSON marshals the Matrix into a pretty-printed JSON string
func (tm *Matrix) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(tm, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal test matrix to JSON: %w", err)
	}

	return string(bytes), nil
}

// newTest creates a new Test from a directory path
func newTest(path string) Test {
	name := filepath.Base(path)

	forwardSlashPath := filepath.ToSlash(path)
	var testType string
	switch {
	case strings.Contains(forwardSlashPath, "/plugin/"):
		testType = "plugin"
	case strings.Contains(forwardSlashPath, "/sdkprovider/"):
		testType = "sdk"
	default:
		testType = "generic"
	}

	return Test{
		Path: path,
		Name: name,
		Type: testType,
	}
}

// pathToUniqueName converts a directory path to a globally unique and stable test name
// for unambiguous identification. Uses the path relative to the provided root.
func pathToUniqueName(root, path string) string {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}

	relPath = filepath.ToSlash(relPath)

	switch {
	case strings.HasPrefix(relPath, "plugin/"):
		return "plugin-" + strings.ReplaceAll(strings.TrimPrefix(relPath, "plugin/"), "/", "-")
	case strings.HasPrefix(relPath, "sdkprovider/"):
		return "sdk-" + strings.ReplaceAll(strings.TrimPrefix(relPath, "sdkprovider/"), "/", "-")
	default:
		return strings.ReplaceAll(relPath, "/", "-")
	}
}
