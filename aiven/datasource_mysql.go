// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceMySQL() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The MySQL data source provides information about the existing Aiven MySQL service.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenMySQLSchema(), "project", "service_name"),
	}
}
