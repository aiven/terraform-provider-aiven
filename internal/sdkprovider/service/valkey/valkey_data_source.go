package valkey

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceValkey() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: userconfig.Desc("The Valkey data source provides information about the existing Aiven for Valkey service.").
			AvailabilityType(userconfig.Beta).
			Build(),
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(valkeySchema(), "project", "service_name"),
	}
}
