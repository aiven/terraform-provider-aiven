// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLRuleRead,
		Description: "The Opensearch ACL Rule data source provides information about an existing Aiven Opensearch ACL Rule.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenOpensearchACLRuleSchema, "project", "service_name", "username", "index", "permission"),
	}
}
