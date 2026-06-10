// Package awsprovision provides custom create (with wait) for aiven_byoc_aws_provision.
package awsprovision

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// deleteViewInternal deletes a custome cloud enviornmnet nd waits until it reaches the DELETED state or CustomCloudEnvironmentGet returns 404.
func deleteViewInternal(ctx context.Context, client avngen.Client, d adapter.ResourceData, retryDelay time.Duration) error {
	organizationID := d.Get("organization_id").(string)
	customCloudEnvironmentID := d.Get("custom_cloud_environment_id").(string)

	return retry.Do(
		func() error {
			cce, err := client.CustomCloudEnvironmentGet(ctx, organizationID, customCloudEnvironmentID)
			if err != nil {
				if avngen.IsNotFound(err) {
					// The CCE is deleted.
					return nil
				}

				// Do not retry on client errors such as 401.
				// 5xx errors are already retried by the client.
				return retry.Unrecoverable(err)
			}

			if cce.State == byoc.CustomCloudEnvironmentStateTypeDeleted {
				// The CCE is deleted.
				return nil
			}

			if cce.State != byoc.CustomCloudEnvironmentStateTypeDeleting {
				// Note: the CCE cannot be deleted while there are services using it,
				// or while a service deletion is still in progress; in these cases the API returns 409.
				if err = client.CustomCloudEnvironmentDelete(ctx, organizationID, customCloudEnvironmentID); err != nil {
					// We don't know exactly which errors this call might return.
					// If it's a 401, the next GET call will mark it as Unrecoverable.
					return err
				}
			}
			return fmt.Errorf("custom cloud environment %s in state %s, waiting for %q", customCloudEnvironmentID, cce.State, byoc.CustomCloudEnvironmentStateTypeDeleted)
		},
		retry.Context(ctx),
		retry.Delay(retryDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			// We can't tell whether a Conflict is caused by Terraform-managed resources that haven't been removed yet
			// or by unmanaged resources still attached to the CCE. Logging the error provides some visibility in those cases.
			tflog.Info(ctx, fmt.Sprintf("Retrying to delete custome cloud environment %q: %s", customCloudEnvironmentID, err))
		}),
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return deleteViewInternal(ctx, client, d, common.DefaultStateChangeDelay)
}
