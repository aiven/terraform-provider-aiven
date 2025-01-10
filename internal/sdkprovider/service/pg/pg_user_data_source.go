package pg

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourcePGUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(DatasourcePGUserRead),
		Description: "Gets information about an Aiven for PostgreSQL® service user.",
		Schema:      DatasourcePGUserSchema(),
	}
}

func DatasourcePGUserSchema() map[string]*schema.Schema {
	return schemautil.ResourceSchemaAsDatasourceSchema(ResourcePGUserSchema, "project", "service_name", "username")
}

func DatasourcePGUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
	return ResourcePGUserRead(ctx, d, client)
}
