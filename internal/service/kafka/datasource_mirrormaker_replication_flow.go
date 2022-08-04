package kafka

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceMirrorMakerReplicationFlowTopic() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceMirrorMakerReplicationFlowRead,
		Description: "The MirrorMaker 2 Replication Flow data source provides information about the existing " +
			"MirrorMaker 2 Replication Flow on Aiven Cloud.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenMirrorMakerReplicationFlowSchema,
			"project", "service_name", "source_cluster", "target_cluster"),
	}
}

func datasourceMirrorMakerReplicationFlowRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(ctx, d, m)
}
