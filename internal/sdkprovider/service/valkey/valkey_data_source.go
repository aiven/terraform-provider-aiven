package valkey

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceValkey() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Valkey service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(valkeySchema(), "project", "service_name"),
	}
}
