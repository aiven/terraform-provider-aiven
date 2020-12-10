package aiven

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceMirrorMakerReplicationFlowTopic() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceMirrorMakerReplicationFlowRead,
		Schema: resourceSchemaAsDatasourceSchema(
			aivenMirrorMakerReplicationFlowSchema,
			"project", "service_name", "source_cluster", "target_cluster"),
	}
}

func datasourceMirrorMakerReplicationFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)

	d.SetId(buildResourceID(projectName, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(ctx, d, m)
}
