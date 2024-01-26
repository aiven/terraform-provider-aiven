package dragonfly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceDragonfly() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Dragonfly data source provides information about the existing Aiven Dragonfly service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(dragonflySchema(), "project", "service_name"),
	}
}
