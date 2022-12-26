package influxdb

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The InfluxDB data source provides information about the existing Aiven InfluxDB service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(influxDBSchema(), "project", "service_name"),
	}
}
