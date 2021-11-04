// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectionPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceConnectionPoolRead,
		Description: "The Connection Pool data source provides information about the existing Aiven Connection Pool.",
		Schema: resourceSchemaAsDatasourceSchema(aivenConnectionPoolSchema,
			"project", "service_name", "pool_name"),
	}
}

func datasourceConnectionPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	poolName := d.Get("pool_name").(string)

	pools, err := client.ConnectionPools.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, pool := range pools {
		if pool.PoolName == poolName {
			d.SetId(buildResourceID(projectName, serviceName, poolName))
			return resourceConnectionPoolRead(ctx, d, m)
		}
	}

	return diag.Errorf("connection pool %s/%s/%s not found",
		projectName, serviceName, poolName)
}
