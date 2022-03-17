package clickhouse

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceClickhouseUserRead,
		Description: "The Clickhouse User data source provides information about the existing Aiven Clickhouse User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenClickhouseUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ClickhouseUser.List(projectName, serviceName)
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
