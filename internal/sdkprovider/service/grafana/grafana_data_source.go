package grafana

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceGrafana() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Grafana® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(grafanaSchema(), "project", "service_name"),
	}
}
