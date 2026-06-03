package projectvpc

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	DataSourceOptions.Read = datasourceReadView
}

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(_ adapter.ResourceData, dto map[string]any) error {
		dto["peering_connections"] = []any{}
		return nil
	}
}

func refreshStateWaiter(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	vpcID := d.Get("project_vpc_id").(string)

	return retry.Do(
		func() error {
			rsp, err := client.VpcGet(ctx, project, vpcID)
			if err != nil {
				if avngen.IsNotFound(err) {
					// The resource may not be available immediately after creation; retry.
					return err
				}

				// Do not retry on client errors such as 401.
				// 5xx errors are already retried by the client.
				return retry.Unrecoverable(err)
			}

			if rsp.State == vpc.VpcStateTypeActive {
				return nil
			}

			return fmt.Errorf("project VPC %s in state %s, waiting for ACTIVE", vpcID, rsp.State)
		},
		retry.Context(ctx),
		retry.Delay(common.DefaultStateChangeDelay),
		retry.LastErrorOnly(true),
	)
}

// deleteViewInternal deletes an Aiven project VPC and waits until it reaches the DELETED state or VpcGet returns 404.
func deleteViewInternal(ctx context.Context, client avngen.Client, d adapter.ResourceData, retryDelay time.Duration) error {
	project := d.Get("project").(string)
	vpcID := d.Get("project_vpc_id").(string)

	return retry.Do(
		func() error {
			rspGet, err := client.VpcGet(ctx, project, vpcID)
			if avngen.IsNotFound(err) {
				// The VPC is fully deleted.
				return nil
			}

			if err != nil {
				// Do not retry on client errors such as 401.
				// 5xx errors are already retried by the client.
				return retry.Unrecoverable(err)
			}

			if rspGet.State == vpc.VpcStateTypeDeleted {
				// The VPC is soft-deleted.
				return nil
			}

			if rspGet.State != vpc.VpcStateTypeDeleting {
				// Note: the VPC cannot be deleted while services are migrating from it,
				// or while a service deletion is still in progress; in these cases the API returns 409.
				_, err = client.VpcDelete(ctx, project, vpcID)
				if err != nil {
					// We don't know exactly which errors this call might return.
					// If it's a 401, the next GET call will mark it as Unrecoverable.
					return err
				}
			}
			return fmt.Errorf("project VPC %s in state %q, waiting for %q", vpcID, rspGet.State, vpc.VpcStateTypeDeleted)
		},
		retry.Context(ctx),
		retry.Delay(retryDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			// We can't tell whether a Conflict is caused by Terraform-managed resources that haven't been removed yet
			// or by unmanaged resources still attached to the VPC. Logging the error provides some visibility in those cases.
			tflog.Info(ctx, fmt.Sprintf("Retrying to delete project VPC %s: %s", vpcID, err))
		}),
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return deleteViewInternal(ctx, client, d, common.DefaultStateChangeDelay)
}

func datasourceReadView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	vpcID := d.Get("vpc_id").(string)
	if vpcID != "" {
		project, vpcID, err := schemautil.SplitResourceID2(vpcID)
		if err != nil {
			return err
		}
		if err := d.Set("project", project); err != nil {
			return err
		}
		if err := d.Set("project_vpc_id", vpcID); err != nil {
			return err
		}
	}
	return readView(ctx, client, d)
}
