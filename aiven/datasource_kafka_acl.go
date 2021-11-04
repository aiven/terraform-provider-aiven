// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaACLRead,
		Description: "The Data Source Kafka ACL data source provides information about the existing Aiven Kafka ACL for a Kafka service.",

		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaACLSchema,
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
			d.SetId(buildResourceID(projectName, serviceName, acl.ID))
			return resourceKafkaACLRead(ctx, d, m)
		}
	}

	return diag.Errorf("KafkaACL %s/%s not found", topic, userName)
}
