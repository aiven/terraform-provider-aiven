package changelog

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	vtool "github.com/aiven/terraform-provider-aiven/tools/version"
)

const (
	FileName        = "CHANGELOG.md"
	changelogHeader = "## [MAJOR.MINOR.PATCH] - YYYY-MM-DD"
)

var (
	// ErrInvalidVersionFormat indicates that the provided version format is invalid.
	ErrInvalidVersionFormat = errors.New("invalid version format")

	// ErrVersionAlreadyExists indicates that the version already exists in the changelog.
	ErrVersionAlreadyExists = errors.New("version already exists")

	// ErrTemplateHeaderNotFound indicates that template header is missing in the file.
	ErrTemplateHeaderNotFound = fmt.Errorf("template header '%s' not found", changelogHeader)
)

// Update adds a new version entry to the specified file, inserting it below the template header.
func Update(filePath, version, date string) error {
	if err := vtool.Validate(version); err != nil {
		return ErrInvalidVersionFormat
	}

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(contentBytes)

	// check if the version already exists using a multiline-aware regex.
	versionExistsRegex, err := regexp.Compile(`(?m)^## \[` + regexp.QuoteMeta(version) + `\]`)
	if err != nil {
		return fmt.Errorf("internal error compiling regex: %w", err) // Should not happen
	}
	if versionExistsRegex.MatchString(content) {
		return fmt.Errorf("%w: %s", ErrVersionAlreadyExists, version)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to re-open file for update: %w", err)
	}
	defer file.Close()

	var newContent strings.Builder
	scanner := bufio.NewScanner(file)
	headerFoundAndPatched := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == changelogHeader {
			headerFoundAndPatched = true
			newHeaderLine := fmt.Sprintf("## [%s] - %s", version, date)

			// write the original header line, a blank line, and then the new version header.
			newContent.WriteString(line + "\n")
			newContent.WriteString("\n")
			newContent.WriteString(newHeaderLine + "\n")
		} else {
			// copy all other lines as-is.
			newContent.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file during update: %w", err)
	}

	// if we never found the header, return an error
	if !headerFoundAndPatched {
		return ErrTemplateHeaderNotFound
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}
	if err = os.WriteFile(filePath, []byte(newContent.String()), info.Mode()); err != nil {
		return fmt.Errorf("failed to write updated changelog to file: %w", err)
	}

	slog.Info("Changelog updated successfully with new", "version", version)

	return nil
}

// Extract finds the notes for a specific version and prints them to stdout
func Extract(filePath, version string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	// find the specific version header.
	startRegex, err := regexp.Compile(`^## \[` + regexp.QuoteMeta(version) + `\]`)
	if err != nil {
		return fmt.Errorf("invalid version string for regex: %w", err)
	}
	// find the start of *any* other version section.
	anyVersionRegex := regexp.MustCompile(`^## \[`)

	var contentBuilder strings.Builder
	capturing := false
	versionFound := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if capturing {
			if anyVersionRegex.MatchString(line) {
				break
			}
			contentBuilder.WriteString(line + "\n")
		} else if startRegex.MatchString(line) {
			capturing = true
			versionFound = true
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	if !versionFound {
		return fmt.Errorf("version %q not found in %s", version, filePath)
	}

	trimmedContent := strings.TrimSpace(contentBuilder.String())
	if trimmedContent == "" {
		return fmt.Errorf("no content found for version %q (the version header exists, but the section is empty)", version)
	}

	// print the final content directly to stdout
	fmt.Println(trimmedContent)

	return nil
}
