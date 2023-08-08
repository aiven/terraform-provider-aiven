package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaACLRead,
		Description: "The Data Source Kafka ACL data source provides information about the existing Aiven Kafka ACL for a Kafka service.",

		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaACLSchema,
			"project", "service_name", "topic", "username", "permission"),
	}
}

func datasourceKafkaACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topic := d.Get("topic").(string)
	userName := d.Get("username").(string)
	permission := d.Get("permission").(string)

	acls, err := client.KafkaACLs.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, acl := range acls {
		if acl.Topic == topic && acl.Username == userName && acl.Permission == permission {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, acl.ID))
			return resourceKafkaACLRead(ctx, d, m)
		}
	}

	return diag.Errorf("KafkaACL %s/%s not found", topic, userName)
}
