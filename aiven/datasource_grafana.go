package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceGrafana() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Grafana data source provides information about the existing Aiven Grafana service.",
		Schema:      resourceSchemaAsDatasourceSchema(grafanaSchema(), "project", "service_name"),
	}
}
