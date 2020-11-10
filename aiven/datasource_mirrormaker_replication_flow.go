package aiven

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func datasourceMirrorMakerReplicationFlowTopic() *schema.Resource {
	return &schema.Resource{
		Read: datasourceMirrorMakerReplicationFlowRead,
		Schema: resourceSchemaAsDatasourceSchema(
			aivenMirrorMakerReplicationFlowSchema,
			"project", "service_name", "source_cluster", "target_cluster"),
	}
}

func datasourceMirrorMakerReplicationFlowRead(d *schema.ResourceData, m interface{}) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)

	d.SetId(buildResourceID(projectName, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(d, m)
}
