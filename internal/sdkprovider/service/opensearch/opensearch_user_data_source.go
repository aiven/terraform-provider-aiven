package opensearch

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOpenSearchUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "Gets information about an Aiven for OpenSearch® service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenOpenSearchUserSchema,
			"project", "service_name", "username"),
	}
}
