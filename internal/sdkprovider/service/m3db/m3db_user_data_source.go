package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceM3DBUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "Gets information about an Aiven for M3DB service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenM3DBUserSchema,
			"project", "service_name", "username"),
	}
}
