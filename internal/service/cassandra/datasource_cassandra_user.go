package cassandra

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceCassandraUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The Cassandra User data source provides information about the existing Aiven Cassandra User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenCassandraUserSchema,
			"project", "service_name", "username"),
	}
}
