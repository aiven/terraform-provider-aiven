package redis

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceRedisUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceRedisUserRead),
		Description: "The Redis User data source provides information about the existing Aiven Redis User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenRedisUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceRedisUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	svc, err := client.ServiceGet(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range svc.Users {
		if u.Username == userName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
			return resourceRedisUserRead(ctx, d, client)
		}
	}

	return diag.Errorf("redis user %s/%s/%s not found",
		projectName, serviceName, userName)
}
