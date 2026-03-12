// Package staticip provides custom create (with wait) and delete (list, dissociate if available, then StaticIPDelete) for aiven_static_ip.
package staticip

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/avast/retry-go"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Delete = deleteWithDissociate
}

func refreshStateWaiter(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	staticIPAddressID := d.Get("static_ip_address_id").(string)
	return retry.Do(
		func() error {
			list, err := client.StaticIPList(ctx, project)
			if err != nil {
				return retry.Unrecoverable(err)
			}
			for _, sip := range list {
				if sip.StaticIPAddressId != d.Get("static_ip_address_id").(string) {
					continue
				}
				if sip.State == staticip.StaticIPStateTypeCreated {
					return nil
				}
				return fmt.Errorf("static ip %s in state %s, waiting for created", staticIPAddressID, sip.State)
			}
			return fmt.Errorf("static ip %s not found in project", staticIPAddressID)
		},
		retry.Context(ctx),
		retry.Delay(common.DefaultStateChangeDelay),
		retry.LastErrorOnly(true),
	)
}

// deleteWithDissociate first attempts to dissociate the static IP before deleting it.
// Any errors from the dissociation call are ignored, as the static IP will be deleted regardless.
// This replaces the previous approach, which checked the IP state before attempting dissociation.
func deleteWithDissociate(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	_, _ = client.ProjectStaticIPDissociate(ctx, d.Get("project").(string), d.Get("static_ip_address_id").(string))
	return deleteView(ctx, client, d)
}
