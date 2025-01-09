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
		ReadContext: common.WithGenClient(ReadDatasourcePGUser),
		Description: "Gets information about an Aiven for PostgreSQL® service user.",
		Schema:      SchemaDatasourcePGUser(),
	}
}

func SchemaDatasourcePGUser() map[string]*schema.Schema {
	return schemautil.ResourceSchemaAsDatasourceSchema(SchemaResourcePGUser, "project", "service_name", "username")
}

func ReadDatasourcePGUser(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
	return ReadResourcePGUser(ctx, d, client)
}
