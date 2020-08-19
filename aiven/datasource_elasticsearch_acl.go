// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceElasticsearchACL() *schema.Resource {
	return &schema.Resource{
		Read: datasourceElasticsearchACLRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenElasticsearchACLSchema,
			"project", "service_name"),
	}
}

func datasourceElasticsearchACLRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ElasticsearchACLs.Get(projectName, serviceName)
	if err != nil {
		return err
	}

	if acl != nil {
		d.SetId(buildResourceID(projectName, serviceName))

		return resourceElasticsearchACLRead(d, m)
	}

	return fmt.Errorf("elasticsearch acl %s/%s not found",
		projectName, serviceName)
}
