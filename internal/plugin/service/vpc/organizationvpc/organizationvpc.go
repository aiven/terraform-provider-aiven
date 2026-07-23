package organizationvpc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	orgvpc "github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

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
