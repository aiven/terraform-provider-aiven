package influxdb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: deprecationMessage,
		ReadContext:        schemautil.DatasourceServiceRead,
		Description:        "The InfluxDB data source provides information about the existing Aiven InfluxDB service.",
		Schema:             schemautil.ResourceSchemaAsDatasourceSchema(influxDBSchema(), "project", "service_name"),
	}
}
