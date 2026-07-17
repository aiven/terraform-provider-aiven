package azureprivatelink

import (
	"context"
	"net/http"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestDeleteView(t *testing.T) {
	newResourceData := func(t *testing.T) adapter.ResourceData {
		t.Helper()

		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestState(map[string]any{
				"id":           "example-project/example-service",
				"project":      "example-project",
				"service_name": "example-service",
			}),
		)
		require.NoError(t, err)
		return d
	}

	t.Run("waits until missing", func(t *testing.T) {
		ctx := context.Background()
		client := avngen.NewMockClient(t)
		d := newResourceData(t)

		client.EXPECT().
			ServicePrivatelinkAzureDelete(ctx, "example-project", "example-service").
			Return(&privatelink.ServicePrivatelinkAzureDeleteOut{
				State: privatelink.ServicePrivatelinkAzureStateTypeDeleting,
			}, nil).
			Once()
		client.EXPECT().
			ServicePrivatelinkAzureGet(ctx, "example-project", "example-service").
			Return(&privatelink.ServicePrivatelinkAzureGetOut{
				State: privatelink.ServicePrivatelinkAzureStateTypeDeleting,
			}, nil).
			Once()
		client.EXPECT().
			ServicePrivatelinkAzureGet(ctx, "example-project", "example-service").
			Return(nil, avngen.Error{Status: http.StatusNotFound}).
			Once()

		require.NoError(t, deleteViewInternal(ctx, client, d, time.Millisecond))
	})

	t.Run("treats missing as deleted", func(t *testing.T) {
		ctx := context.Background()
		client := avngen.NewMockClient(t)
		d := newResourceData(t)

		client.EXPECT().
			ServicePrivatelinkAzureDelete(ctx, "example-project", "example-service").
			Return(nil, avngen.Error{Status: http.StatusNotFound}).
			Once()

		require.NoError(t, deleteViewInternal(ctx, client, d, time.Millisecond))
	})
}
