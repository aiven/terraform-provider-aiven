package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpensearchUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The Opensearch User data source provides information about the existing Aiven Cassandra User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenOpensearchUserSchema,
			"project", "service_name", "username"),
	}
}
