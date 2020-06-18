package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceRedis() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(redisSchema(), "project", "service_name"),
	}
}
