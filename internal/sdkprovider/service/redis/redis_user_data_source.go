package redis

import (
	"context"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
=======

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceRedisUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRedisUserRead,
		Description: "The Redis User data source provides information about the existing Aiven Redis User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenRedisUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceRedisUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range list {
		if u.Username == userName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
			return resourceRedisUserRead(ctx, d, m)
		}
	}

	return diag.Errorf("redis user %s/%s/%s not found",
		projectName, serviceName, userName)
}
