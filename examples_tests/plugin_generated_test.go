//go:build all || examples

package examples

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/aiven/terraform-provider-aiven/internal/acctest"
)

const (
	examplesDir       = "../examples"
	definitionsDir    = "../definitions"
	definitionsPrefix = "aiven_"
)

type definition struct {
	Resource *struct {
		DisableExample bool `yaml:"disableExample"`
	} `yaml:"resource"`
	Datasource *struct {
		DisableExample bool `yaml:"disableExample"`
	} `yaml:"datasource"`
}

// knownAPICallers returns a list of resources that make API calls during the plan.
func knownAPICallers() []string {
	return []string{
		"aiven_kafka_topic", // Validates "topic already exists" during the plan.
	}
}

// TestAccGeneratedExamples tests the generated examples for the provider.
func TestAccGeneratedExamples(t *testing.T) {
	entries, err := os.ReadDir(definitionsDir)
	require.NoError(t, err)

	var resourceFiles []string
	var datasourceFiles []string
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !strings.HasPrefix(name, definitionsPrefix) {
			continue
		}

		defBytes, err := os.ReadFile(filepath.Join(definitionsDir, entry.Name()))
		require.NoError(t, err)

		var def definition
		require.NoError(t, yaml.Unmarshal(defBytes, &def))

		if def.Resource != nil && !def.Resource.DisableExample {
			resourceFiles = append(resourceFiles, filepath.Join(examplesDir, "resources", name, "resource.tf"))
		}

		if def.Datasource != nil && !def.Datasource.DisableExample {
			datasourceFiles = append(datasourceFiles, filepath.Join(examplesDir, "data-sources", name, "data-source.tf"))
		}
	}

	// Resources usually don't make any API calls during the plan.
	// If any does, this is the place to fix the test.
	for _, f := range resourceFiles {
		aivenName := filepath.Base(filepath.Dir(f))
		t.Run(aivenName, func(t *testing.T) {
			config, err := os.ReadFile(f)
			require.NoError(t, err)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.NoopProviderServer(),
				IsUnitTest:               true,
				Steps: []resource.TestStep{
					{
						PlanOnly:           true,
						Config:             string(config),
						ExpectNonEmptyPlan: true,
						ExpectError: func() *regexp.Regexp {
							if slices.Contains(knownAPICallers(), aivenName) {
								return acctest.ErrNoopErrorRegex
							}
							return nil
						}(),
					},
				},
			})
		})
	}

	// Datasources are invoked during the plan, so we expect an error from the noop client.
	// This means the validation has passed, which is what we want.
	for _, f := range datasourceFiles {
		t.Run(f, func(t *testing.T) {
			config, err := os.ReadFile(f)
			require.NoError(t, err)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.NoopProviderServer(),
				IsUnitTest:               true,
				Steps: []resource.TestStep{
					{
						PlanOnly:    true,
						Config:      string(config),
						ExpectError: acctest.ErrNoopErrorRegex,
					},
				},
			})
		})
	}
}
