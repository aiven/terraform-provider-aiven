package apiconvert

import (
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// props is a function that returns a map of properties from a given schema type and node name.
func props(st userconfig.SchemaType, n string) (map[string]interface{}, error) {
	rm, err := userconfig.CachedRepresentationMap(st)
	if err != nil {
		return nil, err
	}

	s, ok := rm[n]
	if !ok {
		return nil, fmt.Errorf("no schema found for %s (type %d)", n, st)
	}

	as, ok := s.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema %s (type %d) is not a map", n, st)
	}

	p, ok := as["properties"]
	if !ok {
		return nil, fmt.Errorf("no properties found for %s (type %d)", n, st)
	}

	ap, ok := p.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("properties of schema %s (type %d) are not a map", n, st)
	}

	return ap, nil
}
