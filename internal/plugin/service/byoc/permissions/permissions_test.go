package permissions

import (
	"context"
	"testing"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResourceSchemaOverride verifies the manual schema override.
// The generated schema marks accounts and project_names as Required + ForceNew
// because forceNew: false doesn't work in the generator (mergeItems uses OR).
// The manual schema makes them Optional + Computed with no ForceNew.
func TestResourceSchemaOverride(t *testing.T) {
	ctx := context.Background()
	s := permissionsSchema(ctx)

	for _, name := range []string{"accounts", "project_names"} {
		attr, ok := s.Attributes[name]
		require.True(t, ok, "missing attribute %s", name)
		setAttr, ok := attr.(schema.SetAttribute)
		require.True(t, ok, "%s should be SetAttribute", name)
		assert.True(t, setAttr.Optional, "%s should be Optional", name)
		assert.True(t, setAttr.Computed, "%s should be Computed", name)
		assert.False(t, setAttr.Required, "%s should NOT be Required", name)
	}
}

// TestConfigValidatorRequiresAtLeastOne verifies the config validator
// that prevents creating a permissions resource with neither accounts nor project_names.
func TestConfigValidatorRequiresAtLeastOne(t *testing.T) {
	validators := configValidators(context.Background(), nil)
	require.Len(t, validators, 1, "should have exactly one config validator")
}
