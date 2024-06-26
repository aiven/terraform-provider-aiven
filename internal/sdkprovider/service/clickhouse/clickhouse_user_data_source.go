package clickhouse

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceClickhouseUserRead,
		Description: "Gets information about a ClickHouse user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenClickhouseUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ClickhouseUser.List(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range list.Users {
		if u.Name == userName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.UUID))
			return resourceClickhouseUserRead(ctx, d, m)
		}
	}

	return diag.Errorf("clickhouse user %s/%s/%s not found",
		projectName, serviceName, userName)
}
