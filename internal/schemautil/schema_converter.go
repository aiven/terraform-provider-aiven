package schemautil

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSchemaAsDatasourceSchema(d map[string]*schema.Schema, required ...string) map[string]*schema.Schema {
	s := make(map[string]*schema.Schema)

outer:
	for k, v := range d {
		// skip write-only fields in data sources
		if v.WriteOnly {
			continue
		}

		// skip fields that require write-only fields (they only make sense together)
		for _, rw := range v.RequiredWith {
			if d[rw].WriteOnly {
				continue outer
			}
		}

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
