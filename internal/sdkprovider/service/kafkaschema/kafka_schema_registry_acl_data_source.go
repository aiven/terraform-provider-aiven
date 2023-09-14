package kafkaschema

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaSchemaRegistryACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaSchemaRegistryACLRead,
		Description: "The Data Source Kafka Schema Registry ACL data source provides information about the existing Aiven Kafka Schema Registry ACL for a Kafka service.",

		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchemaRegistryACLSchema,
			"project", "service_name", "resource", "username", "permission"),
	}
}

func datasourceKafkaSchemaRegistryACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	resource := d.Get("resource").(string)
	userName := d.Get("username").(string)
	permission := d.Get("permission").(string)

	acls, err := client.KafkaSchemaRegistryACLs.List(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, acl := range acls {
		if acl.Resource == resource && acl.Username == userName && acl.Permission == permission {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, acl.ID))
			return resourceKafkaSchemaRegistryACLRead(ctx, d, m)
		}
	}

	return diag.Errorf("KafkaSchemaRegistryACL %s/%s not found", resource, userName)
}
