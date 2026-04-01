package securitypluginconfig

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/opensearch"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Update = updateView
	ResourceOptions.Delete = deleteView
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("admin_password") {
		_, err := client.ServiceOpenSearchSecurityReset(ctx,
			d.Get("project").(string),
			d.Get("service_name").(string),
			&opensearch.ServiceOpenSearchSecurityResetIn{
				AdminPassword: d.GetState("admin_password").(string),
				NewPassword:   d.Get("admin_password").(string),
			},
		)
		return err
	}
	return nil
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	adapter.AddWarning(
		ctx,
		"OpenSearch Security Plugin cannot be disabled",
		"Once enabled, the OpenSearch Security Plugin remains active even after this resource is deleted.",
	)
	return nil
}
