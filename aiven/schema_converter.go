// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSchemaAsDatasourceSchema(d map[string]*schema.Schema, required ...string) map[string]*schema.Schema {
	s := make(map[string]*schema.Schema)
	for k, v := range d {
		s[k] = &schema.Schema{
			Type:        v.Type,
			Required:    false,
			Computed:    v.Computed,
			Optional:    true,
			Description: v.Description,
			Sensitive:   v.Sensitive,
			Elem:        v.Elem,
			MaxItems:    v.MaxItems,
			MinItems:    v.MinItems,
		}
	}
	for _, k := range required {
		s[k].Required, s[k].Optional = true, false
	}
	return s
}
