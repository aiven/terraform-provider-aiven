// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetFixedValues(t *testing.T) {
	t.Parallel()

	t.Run("sets and overrides fields", func(t *testing.T) {
		dto := map[string]any{"cloud_provider": "aws", "display_name": "foo"}
		err := SetFixedValues("cloud_provider", "azure", "deployment_model", "standard")(nil, dto)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"cloud_provider":   "azure",
			"deployment_model": "standard",
			"display_name":     "foo",
		}, dto)
	})

	t.Run("odd number of arguments", func(t *testing.T) {
		err := SetFixedValues("cloud_provider")(nil, map[string]any{})
		require.ErrorContains(t, err, "key-value pairs")
	})
}
