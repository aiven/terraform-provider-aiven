// Package staticip provides custom delete (dissociate then delete) for aiven_static_ip.
package staticip

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Delete = deleteWithDissociate
}

// deleteWithDissociate first attempts to dissociate the static IP before deleting it.
// Any errors from the dissociation call are ignored, as the static IP will be deleted regardless.
// This replaces the previous approach, which checked the IP state before attempting dissociation.
func deleteWithDissociate(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	_, _ = client.ProjectStaticIPDissociate(ctx, d.Get("project").(string), d.Get("static_ip_address_id").(string))
	return deleteView(ctx, client, d)
}
