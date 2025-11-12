package test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/agnivade/levenshtein"
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
func GenerateMatrix(root string, slowTestsCSV string, filterServicesCSV string) (*Matrix, error) {
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

	serviceFilters, err := parseFilters(filterServicesCSV, testDirs)
	if err != nil {
		return nil, err
	}

	// partition tests into normal and slow
	matrix := &Matrix{
		Normal: []Test{},
		Slow:   []Test{},
	}
	for dir := range testDirs {
		test := newTest(dir)

		// apply filters if any
		if len(serviceFilters) > 0 && !matchesAnyService(test.Path, serviceFilters) {
			continue
		}

		uniqueName := pathToUniqueName(root, dir)

		if _, isSlow := slowTestsLookup[uniqueName]; isSlow {
			matrix.Slow = append(matrix.Slow, test)
		} else {
			matrix.Normal = append(matrix.Normal, test)
		}
	}

	return matrix, nil
}

// matchesAnyService checks if a test path matches any of the provided filters
func matchesAnyService(path string, filters []string) bool {
	pathLower := strings.ToLower(filepath.ToSlash(path))

	for _, filter := range filters {
		filterLower := strings.ToLower(strings.TrimSpace(filter))
		if filterLower == "" {
			continue
		}

		// extract the last component of the path (e.g., "kafka" from "internal/sdkprovider/service/kafka")
		parts := strings.Split(pathLower, "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]

			// Match if:
			// 1. The last path component starts with the filter (e.g., "kafka" matches "kafka", "kafkatopic", "kafkaschema")
			// 2. OR the last path component equals the filter exactly
			if strings.HasPrefix(lastPart, filterLower) || lastPart == filterLower {
				return true
			}
		}
	}

	return false
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

// parseFilters parses filterCSV and returns corrected (if any typos) filters with fuzzy matching.
// Returns empty list if filterCSV is empty, corrected filters if valid, or error if no matches found.
func parseFilters(filterCSV string, testDirs map[string]struct{}) ([]string, error) {
	// return empty list if no filter provided
	if filterCSV == "" {
		return nil, nil
	}

	// parse into filters by commas or spaces
	var filters []string
	for _, part := range strings.FieldsFunc(filterCSV, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			filters = append(filters, trimmed)
		}
	}

	uniqueServices := make(map[string]struct{})
	for dir := range testDirs {
		test := newTest(dir)
		uniqueServices[strings.ToLower(test.Name)] = struct{}{}
	}

	const maxDistance = 2
	correctedFilters := make([]string, 0, len(filters))

	for _, filter := range filters {
		filterLower := strings.ToLower(filter)

		// try exact or prefix match
		found := false
		for svc := range uniqueServices {
			if svc == filterLower || strings.HasPrefix(svc, filterLower) {
				correctedFilters = append(correctedFilters, filter)
				found = true
				break
			}
		}

		if found {
			continue
		}

		// try fuzzy match with Levenshtein distance
		bestMatch := ""
		bestDistance := maxDistance + 1
		for svc := range uniqueServices {
			if dist := levenshtein.ComputeDistance(filterLower, svc); dist < bestDistance {
				bestDistance = dist
				bestMatch = svc
			}
		}

		if bestDistance <= maxDistance {
			correctedFilters = append(correctedFilters, bestMatch)
		} else {
			// return error if list was provided but no matches found
			availableList := make([]string, 0, len(uniqueServices))
			for svc := range uniqueServices {
				availableList = append(availableList, svc)
			}

			return nil, fmt.Errorf("no match found for filter: %q\nAvailable services: %v", filter, availableList)
		}
	}

	return correctedFilters, nil
}
