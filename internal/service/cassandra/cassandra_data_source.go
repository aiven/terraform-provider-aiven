package cassandra

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceCassandra() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Cassandra data source provides information about the existing Aiven Cassandra service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(cassandraSchema(), "project", "service_name"),
	}
}
