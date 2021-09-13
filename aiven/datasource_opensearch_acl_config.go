// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLConfigRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenOpensearchACLConfigSchema, "project", "service_name"),
	}
}
