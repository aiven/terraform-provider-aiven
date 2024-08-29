package cassandra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceCassandra() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Apache Cassandra® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(cassandraSchema(), "project", "service_name"),
	}
}
