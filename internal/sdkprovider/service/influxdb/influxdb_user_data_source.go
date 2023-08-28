package influxdb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceInfluxDBUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The InfluxDB User data source provides information about the existing Aiven InfluxDB User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenInfluxDBUserSchema,
			"project", "service_name", "username"),
	}
}
