package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpenSearchUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The OpenSearch User data source provides information about the existing Aiven OpenSearch User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenOpenSearchUserSchema,
			"project", "service_name", "username"),
	}
}
