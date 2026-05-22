package projectvpc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/avast/retry-go"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

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

func waitForDeletion(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	vpcID := d.Get("project_vpc_id").(string)

	return retry.Do(
		func() error {
			getRsp, err := client.VpcGet(ctx, project, vpcID)
			if avngen.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return retry.Unrecoverable(err)
			}

			if getRsp.State == vpc.VpcStateTypeDeleted {
				return nil
			}

			return fmt.Errorf("project VPC %s in state %s, waiting for DELETED", vpcID, getRsp.State)
		},
		retry.Context(ctx),
		retry.Delay(common.DefaultStateChangeDelay),
		retry.LastErrorOnly(true),
	)
}
