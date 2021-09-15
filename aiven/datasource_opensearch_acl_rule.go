// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLRuleRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenOpensearchACLRuleSchema, "project", "service_name", "username", "index", "permission"),
	}
}
