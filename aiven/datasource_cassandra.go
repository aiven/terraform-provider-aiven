package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceCassandra() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(cassandraSchema(), "project", "service_name"),
	}
}
