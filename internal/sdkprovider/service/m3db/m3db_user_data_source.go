package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceM3DBUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The M3DB User data source provides information about the existing Aiven M3DB User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenM3DBUserSchema,
			"project", "service_name", "username"),
	}
}
