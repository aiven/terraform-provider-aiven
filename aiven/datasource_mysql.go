package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceMySQL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenMySQLSchema(), "project", "service_name"),
	}
}
