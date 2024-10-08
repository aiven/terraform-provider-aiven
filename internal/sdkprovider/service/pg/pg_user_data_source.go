package pg

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourcePGUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourcePGUserRead,
		Description: "Gets information about an Aiven for PostgreSQL® service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenPGUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourcePGUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range list {
		if u.Username == userName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
			return resourcePGUserRead(ctx, d, m)
		}
	}

	return diag.Errorf("pg user %s/%s/%s not found",
		projectName, serviceName, userName)
}
