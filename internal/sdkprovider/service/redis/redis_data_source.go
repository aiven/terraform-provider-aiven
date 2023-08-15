package redis

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceRedis() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Redis data source provides information about the existing Aiven Redis service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(redisSchema(), "project", "service_name"),
	}
}
