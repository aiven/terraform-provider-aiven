package service_user

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceServiceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceUserRead,
		Description: "The Service User data source provides information about the existing Aiven Service User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenServiceUserSchema,
			"project", "service_name", "username"),
		DeprecationMessage: "Please use service-specific data-sources instead of this one, for example: aiven_kafka_user, aiven_pg_user etc.",
	}
}

func datasourceServiceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

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
			return resourceServiceUserRead(ctx, d, m)
		}
	}

	return diag.Errorf("common user %s/%s/%s not found",
		projectName, serviceName, userName)
}
