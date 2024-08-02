package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// version is the version of the provider.
const version = "test"

func TestProvider(t *testing.T) {
	p := Provider(version)

	err := p.InternalValidate()
	assert.NoError(t, err)

	// Validates deprecations
	sources := []map[string]*schema.Resource{p.ResourcesMap, p.DataSourcesMap}
	for _, s := range sources {
		for k, r := range s {
			if r.DeprecationMessage == "" {
				continue
			}

			assert.Contains(
				t, strings.ToLower(r.Description), "deprecate",
				"%q must have deprecation message and migration instructions", k,
			)
		}
	}
}

func TestProviderImpl(*testing.T) {
	var _ = Provider(version)
}
