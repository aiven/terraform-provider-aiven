package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

const reportFileName = "PLUGIN_MIGRATION.md"

// genReport writes a table with: resource name, plugin indicator, count to reportFileName.
func genReport() error {
	sdkProvider, err := provider.Provider("foo")
	if err != nil {
		return err
	}

	allRes := make(map[string]int)
	pluginRes := make(map[string]bool)
	for k := range plugin.ResourcesMap() {
		allRes[k]++
		pluginRes[k] = true
	}
	for k := range plugin.DataSourcesMap() {
		allRes[k]++
		pluginRes[k] = true
	}
	for k := range sdkProvider.ResourcesMap {
		allRes[k]++
	}
	for k := range sdkProvider.DataSourcesMap {
		allRes[k]++
	}

	// Write table to reportFileName.
	totalSdk := 0
	totalPlugin := 0
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"#", "RESOURCE NAME", "PLUGIN", "COUNT"})
	for i, k := range sortedKeys(allRes) {
		v := allRes[k]
		var p string
		if pluginRes[k] {
			p = "yes"
			totalPlugin += v
		} else {
			totalSdk += v
		}
		tw.AppendRow(table.Row{i + 1, k, p, v})
	}

	totalCount := totalSdk + totalPlugin
	totalMigrated := fmt.Sprintf("TOTAL MIGRATED %.f%%", 100*float64(totalPlugin)/float64(totalCount))
	tw.AppendFooter(
		table.Row{"", totalMigrated, totalPlugin, totalCount},
	)

	builder := strings.Builder{}
	builder.WriteString("# Plugin Framework Migration Status\n\n")
	builder.WriteString("```\n")
	builder.WriteString(tw.Render())
	builder.WriteString("\n```\n")
	return os.WriteFile(reportFileName, []byte(builder.String()), 0o644)
}
