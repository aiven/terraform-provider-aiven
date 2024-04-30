package thanos

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceThanos() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: userconfig.Desc("Gets information about an Aiven for Thanos® service.").
			AvailabilityType(userconfig.Beta).
			Build(),
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(thanosSchema(), "project", "service_name"),
	}
}
