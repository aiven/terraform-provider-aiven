package projectvpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"

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
	projectVPCID := d.Get("project_vpc_id").(string)

	for {
		rsp, err := client.VpcGet(ctx, project, projectVPCID)
		if err != nil {
			return err
		}
		if rsp.State == vpc.VpcStateTypeActive {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("project VPC %q is in state %q: %w", projectVPCID, rsp.State, ctx.Err())
		case <-time.After(common.DefaultStateChangeDelay):
		}
	}
}

func waitForDeletion(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	projectVPCID := d.Get("project_vpc_id").(string)

	for {
		rsp, err := client.VpcGet(ctx, project, projectVPCID)
		if avngen.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if rsp.State == vpc.VpcStateTypeDeleted {
			return nil
		}

		if rsp.State != vpc.VpcStateTypeDeleting {
			_, err = client.VpcDelete(ctx, project, projectVPCID)
			if avngen.IsNotFound(err) {
				return nil
			}
			var apiErr avngen.Error
			if err != nil && (!errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict) {
				return err
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("project VPC %q is in state %q: %w", projectVPCID, rsp.State, ctx.Err())
		case <-time.After(common.DefaultStateChangeDelay):
		}
	}
}
