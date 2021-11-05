// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceM3DB() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The M3 DB data source provides information about the existing Aiven M3 services.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenM3DBSchema(), "project", "service_name"),
	}
}
