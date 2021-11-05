// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Kafka MirrorMaker data source provides information about the existing Aiven Kafka MirrorMaker 2 service.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenKafkaMirrormakerSchema(), "project", "service_name"),
	}
}
