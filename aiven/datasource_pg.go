// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourcePG() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The PG data source provides information about the existing Aiven PostgreSQL service.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenPGSchema(), "project", "service_name"),
	}
}
