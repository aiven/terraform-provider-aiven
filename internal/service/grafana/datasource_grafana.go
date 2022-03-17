package grafana

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceGrafana() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Grafana data source provides information about the existing Aiven Grafana service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(grafanaSchema(), "project", "service_name"),
	}
}
