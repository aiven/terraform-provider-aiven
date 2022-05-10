package influxdb

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceInfluxDBUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The InfluxDB User data source provides information about the existing Aiven InfluxDB User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenInfluxDBUserSchema,
			"project", "service_name", "username"),
	}
}
