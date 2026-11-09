package staticip

import (
	"context"
	"errors"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// The IP is deleted even when it can't be dissociated.
func TestDeleteIgnoresDissociateError(t *testing.T) {
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":                   "my-project/ip-id",
			"project":              "my-project",
			"static_ip_address_id": "ip-id",
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	client := avngen.NewMockClient(t)
	dissociate := client.EXPECT().
		ProjectStaticIPDissociate(ctx, "my-project", "ip-id").
		Return(nil, errors.New("not associated")).
		Once()
	client.EXPECT().
		StaticIPDelete(ctx, "my-project", "ip-id").
		Return(nil, nil).
		Once().
		NotBefore(dissociate)

	require.NoError(t, ResourceOptions.Delete(ctx, client, d))
}
