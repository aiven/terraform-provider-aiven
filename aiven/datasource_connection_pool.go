// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectionPool() *schema.Resource {
	return &schema.Resource{
		Read: datasourceConnectionPoolRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenConnectionPoolSchema,
			"project", "service_name", "pool_name"),
	}
}

func datasourceConnectionPoolRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	poolName := d.Get("pool_name").(string)

	pools, err := client.ConnectionPools.List(projectName, serviceName)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		if pool.PoolName == poolName {
			d.SetId(buildResourceID(projectName, serviceName, poolName))
			return resourceConnectionPoolRead(d, m)
		}
	}

	return fmt.Errorf("connection pool %s/%s/%s not found",
		projectName, serviceName, poolName)
}
