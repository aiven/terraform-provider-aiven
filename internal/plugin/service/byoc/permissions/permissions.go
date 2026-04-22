package permissions

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Delete = deletePermissions
}

func deletePermissions(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return client.CustomCloudEnvironmentPermissionsSet(
		ctx,
		d.Get("organization_id").(string),
		d.Get("custom_cloud_environment_id").(string),
		&byoc.CustomCloudEnvironmentPermissionsSetIn{
			Accounts: []string{},
			Projects: []string{},
		},
	)
}
