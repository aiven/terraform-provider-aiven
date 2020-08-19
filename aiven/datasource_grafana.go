package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceGrafana() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(grafanaSchema(), "project", "service_name"),
	}
}
