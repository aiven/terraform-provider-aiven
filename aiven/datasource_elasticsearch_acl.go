// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceElasticsearchACL() *schema.Resource {
	return &schema.Resource{
		Read:   resourceElasticsearchACLRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenElasticsearchACLSchema, "project", "service_name"),
	}
}
