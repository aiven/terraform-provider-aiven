package projectvpc

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

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
