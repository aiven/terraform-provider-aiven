package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	reportFileName   = "PLUGIN_MIGRATION.md"
	providerGoSuffix = "provider.go"
)

// genReport recursively searches for all `*provider.go` files,
// extracts all matches of "aiven_*" strings,
// and writes a table with: resource name, plugin indicator, count to reportFileName.
func genReport() error {
	// Find all *provider.go files
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, providerGoSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	type entry struct {
		name   string
		plugin bool
		count  int
	}

	// Finds all matches in dictionary keys.
	// `"aiven_foo":` or `foo["aiven_foo"]`
	reAivenMapKey := regexp.MustCompile(`"aiven_[a-z0-9_]+"[:\]]`)
	entriesMap := make(map[string]entry)
	entriesCount := make(map[bool]int)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		matches := reAivenMapKey.FindAll(content, -1)
		if matches == nil {
			continue
		}

		plugin := strings.Contains(file, "plugin")
		for _, m := range matches {
			// m includes quotes, remove them
			name := strings.Trim(string(m), `":]`)
			item := entriesMap[name]
			item.name = name
			item.plugin = plugin
			item.count++
			entriesMap[name] = item
			entriesCount[plugin]++
		}
	}

	// Sort entries alphabetically by name.
	// It used to be "plugin first", but the git diff was hard to read.
	entries := slices.Collect(maps.Values(entriesMap))
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Write table to reportFileName.
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"#", "RESOURCE NAME", "PLUGIN", "COUNT"})
	for i, e := range entries {
		plugin := ""
		if e.plugin {
			plugin = "yes"
		}
		tw.AppendRow(table.Row{i + 1, e.name, plugin, e.count})
	}

	sdkv2Count := entriesCount[false]
	pluginCount := entriesCount[true]
	totalCount := sdkv2Count + pluginCount
	totalMigrated := fmt.Sprintf("TOTAL MIGRATED %.f%%", 100*float64(pluginCount)/float64(totalCount))
	tw.AppendFooter(
		table.Row{"", totalMigrated, pluginCount, totalCount},
	)

	builder := strings.Builder{}
	builder.WriteString("# Plugin Framework Migration Status\n\n")
	builder.WriteString("```\n")
	builder.WriteString(tw.Render())
	builder.WriteString("\n```\n")
	return os.WriteFile(reportFileName, []byte(builder.String()), 0o644)
}
