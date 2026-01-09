package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

const reportFileName = "PLUGIN_MIGRATION.md"

func main() {
	if err := exec(); err != nil {
		log.Fatal().Err(err).Msg("failed to generate report")
	}
}

func exec() error {
	file, err := os.Create(reportFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return report(file)
}

// report generates a migration status report comparing plugin framework resources
// and data sources with SDK provider resources and data sources.
func report(w io.Writer) error {
	if !util.IsBeta() {
		return fmt.Errorf("please enable beta mode, i.e. set %s=1", util.AivenEnableBeta)
	}

	sdkProvider, err := provider.Provider("foo")
	if err != nil {
		return err
	}

	// Collect all resources and data sources, marking which ones use the plugin framework
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

	// Build table with migration status
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

	_, err = fmt.Fprintf(w, "# Plugin Framework Migration Status\n\n```\n%s\n```\n", tw.Render())
	return err
}

// sortedKeys sorts the keys alphabetically
func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	return keys
}
