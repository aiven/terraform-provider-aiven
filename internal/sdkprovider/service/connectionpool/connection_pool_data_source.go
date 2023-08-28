package connectionpool

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceConnectionPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceConnectionPoolRead,
		Description: "The Connection Pool data source provides information about the existing Aiven Connection Pool.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenConnectionPoolSchema,
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
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, poolName))
			return resourceConnectionPoolRead(ctx, d, m)
		}
	}

	return diag.Errorf("connection pool %s/%s/%s not found",
		projectName, serviceName, poolName)
}
