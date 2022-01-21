// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceClickhouse() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Clickhouse data source provides information about the existing Aiven Clickhouse service.",
		Schema:      resourceSchemaAsDatasourceSchema(clickhouseSchema(), "project", "service_name"),
	}
}
