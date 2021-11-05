// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceCassandra() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Cassandra data source provides information about the existing Aiven Cassandra service.",
		Schema:      resourceSchemaAsDatasourceSchema(cassandraSchema(), "project", "service_name"),
	}
}
