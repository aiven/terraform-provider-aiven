package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(influxDBSchema(), "project", "service_name"),
	}
}
