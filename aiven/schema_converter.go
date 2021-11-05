// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSchemaAsDatasourceSchema(d map[string]*schema.Schema, required ...string) map[string]*schema.Schema {
	s := make(map[string]*schema.Schema)
	for k, v := range d {
		s[k] = &schema.Schema{
			Type:        v.Type,
			Computed:    true,
			Description: v.Description,
			Sensitive:   v.Sensitive,
			Elem:        v.Elem,
		}
	}
	for _, k := range required {
		s[k].Required = true
		s[k].Optional = false
		s[k].Computed = false
	}
	return s
}
