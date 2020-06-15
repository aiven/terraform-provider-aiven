package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaMirrormakerSchema(), "project", "service_name"),
	}
}
