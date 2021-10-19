package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceRedis() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Redis data source provides information about the existing Aiven Redis service.",
		Schema:      resourceSchemaAsDatasourceSchema(redisSchema(), "project", "service_name"),
	}
}
