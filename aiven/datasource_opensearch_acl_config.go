// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceElasticsearchACLConfigRead,
		Description: "The Opensearch ACL Config data source provides information about an existing Aiven Opensearch ACL Config.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenOpensearchACLConfigSchema, "project", "service_name"),
	}
}
