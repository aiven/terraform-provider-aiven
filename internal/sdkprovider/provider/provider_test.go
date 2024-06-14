package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// version is the version of the provider.
const version = "test"

func TestProvider(t *testing.T) {
	p, err := Provider(version)
	require.NoError(t, err)

	err = p.InternalValidate()
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
				"%q must have deprecation message and migration instructions in the description", k,
			)
		}
	}
}

func TestProviderImpl(t *testing.T) {
	var _, err = Provider(version)
	assert.NoError(t, err)
}
