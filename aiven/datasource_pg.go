package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourcePG() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenPGSchema(), "project", "service_name"),
	}
}
