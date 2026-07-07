package organizationvpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	orgvpc "github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

var errOrganizationVPCDeletePending = errors.New("organization VPC delete is pending")

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		dto["clouds"] = []any{
			map[string]any{
				"cloud_name":   d.Get("cloud_name"),
				"network_cidr": d.Get("network_cidr"),
			},
		}
		// The API requires peering_connections in create requests.
		// Terraform manages those connections with separate resources.
		dto["peering_connections"] = []any{}
		return nil
	}
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(_ adapter.ResourceData, dto map[string]any) error {
		clouds, ok := dto["clouds"].([]any)
		if !ok {
			return fmt.Errorf("expected `clouds` to be a list, got %T", dto["clouds"])
		}
		if len(clouds) != 1 {
			return fmt.Errorf("expected exactly 1 cloud, got %d", len(clouds))
		}

		cloud, ok := clouds[0].(map[string]any)
		if !ok {
			return fmt.Errorf("expected cloud to be an object, got %T", clouds[0])
		}

		dto["cloud_name"] = cloud["cloud_name"]
		dto["network_cidr"] = cloud["network_cidr"]
		delete(dto, "clouds")
		return nil
	}
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	req := new(orgvpc.OrganizationVpcCreateIn)
	if err := d.Expand(req, expandModifier(ctx, client)); err != nil {
		return err
	}

	rsp, err := client.OrganizationVpcCreate(ctx, d.Get("organization_id").(string), req)
	if err != nil {
		return err
	}
	return d.Flatten(rsp, flattenModifier(ctx, client))
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return deleteViewInternal(ctx, client, d, common.DefaultStateChangeDelay)
}

// deleteViewInternal deletes an Aiven organization VPC and waits until it reaches the DELETED state or the API returns 404.
func deleteViewInternal(ctx context.Context, client avngen.Client, d adapter.ResourceData, retryDelay time.Duration) error {
	orgID := d.Get("organization_id").(string)
	vpcID := d.Get("organization_vpc_id").(string)

	err := retry.Do(
		func() error {
			rspGet, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
			if err != nil {
				return err
			}

			if rspGet.State == orgvpc.OrganizationVpcStateTypeDeleted {
				// The organization VPC is soft-deleted.
				return nil
			}

			if rspGet.State != orgvpc.OrganizationVpcStateTypeDeleting {
				// Note: the organization VPC cannot be deleted while services are migrating from it,
				// or while a service deletion is still in progress; in these cases the API returns 409.
				_, err = client.OrganizationVpcDelete(ctx, orgID, vpcID)
				if err != nil {
					return err
				}
			}
			return fmt.Errorf("%w: organization VPC %s in state %q, waiting for %q", errOrganizationVPCDeletePending, vpcID, rspGet.State, orgvpc.OrganizationVpcStateTypeDeleted)
		},
		retry.Context(ctx),
		retry.Delay(retryDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			if errors.Is(err, errOrganizationVPCDeletePending) {
				return true
			}

			var apiErr avngen.Error
			return errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict
		}),
		retry.OnRetry(func(n uint, err error) {
			// We can't tell whether a Conflict is caused by Terraform-managed resources that haven't been removed yet
			// or by unmanaged resources still attached to the VPC. Logging the error provides some visibility in those cases.
			tflog.Info(ctx, fmt.Sprintf("Retrying to delete organization VPC %s: %s", vpcID, err))
		}),
	)
	if avngen.IsNotFound(err) {
		return nil
	}
	return err
}
