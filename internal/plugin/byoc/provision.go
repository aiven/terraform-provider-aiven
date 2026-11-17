package byoc

import (
	"context"
	"fmt"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// unprovisioned are the states of an environment that provisioning hasn't activated yet.
var unprovisioned = []byoc.CustomCloudEnvironmentStateType{
	byoc.CustomCloudEnvironmentStateTypeDraft,
	byoc.CustomCloudEnvironmentStateTypeValidating,
	byoc.CustomCloudEnvironmentStateTypeCreating,
	byoc.CustomCloudEnvironmentStateTypeCreationFailed,
}

// DeleteProvisioned skips the delete of an environment that provisioning hasn't activated, e.g., `creation_failed`.
// The entity resource owns it, and Terraform replaces a provision resource tainted by a failed provisioning.
// An activated one is deleted here, before the customer-side infrastructure it depends on.
func DeleteProvisioned(
	del func(context.Context, avngen.Client, adapter.ResourceData) error,
) func(context.Context, avngen.Client, adapter.ResourceData) error {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		env, err := client.CustomCloudEnvironmentGet(ctx, d.Get("organization_id").(string), d.Get("custom_cloud_environment_id").(string))
		if err != nil {
			return err
		}
		if slices.Contains(unprovisioned, env.State) {
			return fmt.Errorf("%w: the custom cloud environment is %q, its entity resource deletes it", adapter.ErrDeleteSkipped, env.State)
		}
		return del(ctx, client, d)
	}
}
