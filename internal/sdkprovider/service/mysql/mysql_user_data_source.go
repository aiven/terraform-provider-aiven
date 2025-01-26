package mysql

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceMySQLUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "Gets information about an Aiven for MySQL® service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenMySQLUserSchema,
			"project", "service_name", "username"),
	}
}
