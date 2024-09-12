package cassandra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceCassandraUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "Gets information about an Aiven for Apache Cassandra® service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenCassandraUserSchema,
			"project", "service_name", "username"),
	}
}
