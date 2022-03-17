package redis

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceRedis() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Redis data source provides information about the existing Aiven Redis service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(redisSchema(), "project", "service_name"),
	}
}
