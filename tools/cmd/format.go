package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	defaultExtensions   = []string{".go", ".yml", ".yaml", ".md", ".json"}
	defaultExcludeFiles = []string{"definitions/.schema.yml"}
	defaultExcludeDirs  = []string{".git", "vendor"}

	excludeExtensions []string
	excludeFiles      []string
	excludeDirs       []string
	rootDir           string
	workers           int
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Fast formatting for basic whitespace cleanup",
	Long: `Fast formatting tool for basic whitespace cleanup without Docker overhead.

This is a custom Go implementation that performs the core formatting tasks of pre-commit hooks:
- Removes trailing whitespace (spaces and tabs) from lines
- Normalizes file endings (preserves original newline behavior)
- Processes .go, .yml, .yaml, .md, .json files by default
- Concurrent processing for better performance

Ideal for:
- Development workflow and post-generation cleanup
- CI/CD environments where speed is important
- When Docker is not available or desired

For comprehensive validation including YAML validation, Terraform lint, and other checks,
use the full pre-commit Docker-based workflow instead.`,
	RunE: runFormat,
}

func init() {
	formatCmd.Flags().StringSliceVar(&excludeExtensions, "exclude-ext", []string{}, "File extensions to exclude (e.g., .tmp,.bak)")
	formatCmd.Flags().StringSliceVar(&excludeFiles, "exclude-files", defaultExcludeFiles, "Specific files to exclude")
	formatCmd.Flags().StringSliceVar(&excludeDirs, "exclude-dirs", defaultExcludeDirs, "Directories to exclude")
	formatCmd.Flags().StringVar(&rootDir, "root", ".", "Root directory to format (default: current directory)")
	formatCmd.Flags().IntVar(&workers, "workers", runtime.NumCPU(), "Number of concurrent workers")
}

func runFormat(_ *cobra.Command, _ []string) error {
	var (
		targetExtSet   = toSet(defaultExtensions)
		excludeExtSet  = toSet(excludeExtensions)
		excludeFileSet = toSet(excludeFiles)
		excludeDirSet  = toSet(excludeDirs)

		filesToFormat []string
	)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if _, ok := excludeDirSet[info.Name()]; ok {
				return filepath.SkipDir
			}
			return nil
		}

		// check excluded files
		if _, ok := excludeFileSet[path]; ok {
			return nil
		}

		// check excluded extensions
		ext := filepath.Ext(path)
		if _, ok := excludeExtSet[ext]; ok {
			return nil
		}

		if _, ok := targetExtSet[ext]; ok {
			filesToFormat = append(filesToFormat, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory %s: %w", rootDir, err)
	}

	if err = processFiles(filesToFormat); err != nil {
		return err
	}

	fmt.Printf("formatting completed successfully! Processed %d files.\n", len(filesToFormat))
	return nil
}

func processFiles(files []string) error {
	fileCh := make(chan string, len(files))
	errorCh := make(chan error, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileCh {
				if err := formatFile(file); err != nil {
					select {
					case errorCh <- fmt.Errorf("failed to format %s: %w", file, err):
					default:
					}
				}
			}
		}()
	}

	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	var allErrors []string
	for err := range errorCh {
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to format %d files:\n- %s", len(allErrors), strings.Join(allErrors, "\n- "))
	}

	return nil
}

func formatFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat file: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	if len(content) == 0 {
		return nil
	}

	originalHasNewline := bytes.HasSuffix(content, []byte("\n"))

	lines := bytes.Split(content, []byte("\n"))

	// trim trailing space and tabs
	for i, line := range lines {
		lines[i] = bytes.TrimRight(line, " \t")
	}

	newContent := bytes.Join(lines, []byte("\n"))

	// trim ALL trailing whitespace
	newContent = bytes.TrimRight(newContent, " \t\n")

	// if the original had a newline, add one back
	if originalHasNewline {
		newContent = append(newContent, '\n')
	}

	if !bytes.Equal(content, newContent) {
		return os.WriteFile(path, newContent, fileInfo.Mode())
	}

	return nil
}

// toSet converts a slice to a map
func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}

	return set
}
