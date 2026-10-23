package serviceintegrationendpoint

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestNotFound(t *testing.T) {
	missing := avngen.Error{
		Status:  http.StatusBadRequest,
		Message: "Integration endpoint 'endpoint-id' does not exist",
	}
	invalid := avngen.Error{
		Status:  http.StatusBadRequest,
		Message: "Invalid input",
	}

	newResourceData := func(t *testing.T) adapter.ResourceData {
		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestState(map[string]any{
				"id":          "my-project/endpoint-id",
				"project":     "my-project",
				"endpoint_id": "endpoint-id",
			}),
		)
		require.NoError(t, err)
		return d
	}

	cases := []struct {
		name     string
		err      error
		notFound bool
	}{
		{name: "missing endpoint", err: missing, notFound: true},
		{name: "other bad request", err: invalid, notFound: false},
	}

	for _, c := range cases {
		t.Run("read "+c.name, func(t *testing.T) {
			ctx := context.Background()
			client := avngen.NewMockClient(t)
			client.EXPECT().
				ServiceIntegrationEndpointGet(ctx, "my-project", "endpoint-id", [][2]string{{"include_secrets", "true"}}).
				Return(nil, c.err).
				Once()

			err := ResourceOptions.Read(ctx, client, newResourceData(t))
			require.ErrorIs(t, err, c.err)
			assert.Equal(t, c.notFound, adapter.IsNotFound(err))
		})

		t.Run("delete "+c.name, func(t *testing.T) {
			ctx := context.Background()
			client := avngen.NewMockClient(t)
			client.EXPECT().
				ServiceIntegrationEndpointDelete(ctx, "my-project", "endpoint-id").
				Return(c.err).
				Once()

			err := ResourceOptions.Delete(ctx, client, newResourceData(t))
			require.ErrorIs(t, err, c.err)
			assert.Equal(t, c.notFound, adapter.IsNotFound(err))
		})
	}
}
