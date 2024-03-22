package dragonfly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceDragonfly() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Dragonfly® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(dragonflySchema(), "project", "service_name"),
	}
}
