package alloydbomni

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAlloyDBOmni() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for AlloyDB Omni service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenAlloyDBOmniSchema(), "project", "service_name"),
	}
}
