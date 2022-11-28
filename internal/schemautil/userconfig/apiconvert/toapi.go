package apiconvert

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// resourceDatable is an interface that allows to get the resource data from the schema.
// This is needed to be able to test the conversion functions. See schema.ResourceData for more.
type resourceDatable interface {
	GetOk(string) (interface{}, bool)
	HasChange(string) bool
	IsNewResource() bool
}

// arrayItemToAPI is a function that converts array property of Terraform user configuration schema to API
// compatible format.
func arrayItemToAPI(
	n string,
	fk []string,
	k string,
	v []interface{},
	i map[string]interface{},
	d resourceDatable,
) (interface{}, bool, error) {
	var res []interface{}

	if len(v) == 0 {
		res = []interface{}{}

		return res, false, nil
	}

	fks := strings.Join(fk, ".")

	// TODO: Remove when this is fixed on backend.
	if k == "additional_backup_regions" {
		return res, true, nil
	}

	ii, ok := i["items"].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%s (item): items key not found", fks)
	}

	var iit string

	// If the key has a type suffix, we use it to determine the type of the value.
	if userconfig.IsKeyTyped(k) {
		iit = k[strings.LastIndexByte(k, '_')+1:]

		// Find the one_of item that matches the type.
		if oo, ok := ii["one_of"]; ok {
			ooa, ok := oo.([]interface{})
			if !ok {
				return nil, false, fmt.Errorf("%s (items.one_of): not a slice", fks)
			}

			for i, vn := range ooa {
				vna, ok := vn.(map[string]interface{})
				if !ok {
					return nil, false, fmt.Errorf("%s (items.one_of.%d): not a map", fks, i)
				}

				if ot, ok := vna["type"]; ok && ot == iit {
					ii = vna

					break
				}
			}
		}
	} else {
		// TODO: Remove this statement and the branch below it with the next major version.
		_, ok := ii["one_of"]

		if k == "ip_filter" || (ok && k == "namespaces") {
			iit = "string"
		} else {
			_, aiits, err := userconfig.TerraformTypes(userconfig.SlicedString(ii["type"]))
			if err != nil {
				return nil, false, err
			}

			if len(aiits) > 1 {
				return nil, false, fmt.Errorf("%s (type): multiple types", fks)
			}

			iit = aiits[0]
		}
	}

	for i, vn := range v {
		// We only accept slices there, so we need to nest the value into a slice if the value is of object type.
		if iit == "object" {
			vn = []interface{}{vn}
		}

		vnc, o, err := itemToAPI(n, iit, append(fk, fmt.Sprintf("%d", i)), fmt.Sprintf("%s.%d", k, i), vn, ii, d)
		if err != nil {
			return nil, false, err
		}

		if !o {
			res = append(res, vnc)
		}
	}

	return res, false, nil
}

// objectItemToAPI is a function that converts object property of Terraform user configuration schema to API
// compatible format.
func objectItemToAPI(
	n string,
	fk []string,
	v []interface{},
	i map[string]interface{},
	d resourceDatable,
) (interface{}, bool, error) {
	var res interface{}

	fks := strings.Join(fk, ".")

	fv := v[0]

	fva, ok := fv.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%s: not a map", fks)
	}

	ip, ok := i["properties"].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%s (item): properties key not found", fks)
	}

	if !regexp.MustCompile(`.+\.[0-9]$`).MatchString(fks) {
		fk = append(fk, "0")
	}

	res, err := propsToAPI(n, fk, fva, ip, d)
	if err != nil {
		return nil, false, err
	}

	return res, false, nil
}

// itemToAPI is a function that converts property of Terraform user configuration schema to API compatible format.
func itemToAPI(
	n string,
	t string,
	fk []string,
	k string,
	v interface{},
	i map[string]interface{},
	d resourceDatable,
) (interface{}, bool, error) {
	// TODO: Remove this variable when we use actual types in the schema.
	var err error

	res := v

	fks := strings.Join(fk, ".")

	// We omit the value if has no changes in the Terraform user configuration.
	o := !d.HasChange(fks)

	// TODO: Remove this statement and the branch below it when we use actual types in the schema.
	if va, ok := v.(string); ok && va == "" {
		return res, o, nil
	}

	// Assert the type of the value to match.
	switch t {
	case "boolean":
		// TODO: Uncomment this, and the same below, when we use actual types in the schema.
		// if _, ok := v.(bool); !ok {
		if res, err = strconv.ParseBool(v.(string)); err != nil {
			return nil, false, fmt.Errorf("%s: not a boolean", fks)
		}
	case "integer":
		// if _, ok := v.(int); !ok {
		if res, err = strconv.Atoi(v.(string)); err != nil {
			return nil, false, fmt.Errorf("%s: not an integer", fks)
		}
	case "number":
		// if _, ok := v.(float64); !ok {
		if res, err = strconv.ParseFloat(v.(string), 64); err != nil {
			return nil, false, fmt.Errorf("%s: not a number", fks)
		}
	case "string":
		if _, ok := v.(string); !ok {
			return nil, false, fmt.Errorf("%s: not a string", fks)
		}
	case "array", "object":
		// Arrays and objects are handled separately.

		va, ok := v.([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("%s: not a slice", fks)
		}

		if va == nil || o {
			return nil, true, nil
		}

		if t == "array" {
			return arrayItemToAPI(n, fk, k, va, i, d)
		}

		if len(va) == 0 {
			return nil, true, nil
		}

		return objectItemToAPI(n, fk, va, i, d)
	default:
		return nil, false, fmt.Errorf("%s: unsupported type %s", fks, t)
	}

	return res, o, nil
}

// propsToAPI is a function that converts properties of Terraform user configuration schema to API compatible format.
func propsToAPI(
	n string,
	fk []string,
	tp map[string]interface{},
	p map[string]interface{},
	d resourceDatable,
) (map[string]interface{}, error) {
	res := make(map[string]interface{}, len(tp))

	fks := strings.Join(fk, ".")

	for k, v := range tp {
		k = userconfig.DecodeKey(k)

		rk := k

		// If the key has a suffix, we need to strip it to be able to find the corresponding property in the schema.
		if userconfig.IsKeyTyped(k) {
			rk = k[:strings.LastIndexByte(k, '_')]
		}

		i, ok := p[rk]
		if !ok {
			return nil, fmt.Errorf("%s.%s: key not found", fks, k)
		}

		if i == nil {
			continue
		}

		ia, ok := i.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s.%s: not a map", fks, k)
		}

		// If the property is supposed to be present only during resource's creation,
		// we need to skip it if the resource is being updated.
		if co, ok := ia["create_only"]; ok && co.(bool) && !d.IsNewResource() {
			continue
		}

		_, ats, err := userconfig.TerraformTypes(userconfig.SlicedString(ia["type"]))
		if err != nil {
			return nil, err
		}

		if len(ats) > 1 {
			return nil, fmt.Errorf("%s.%s.type: multiple types", fks, k)
		}

		t := ats[0]

		cv, o, err := itemToAPI(n, t, append(fk, k), k, v, ia, d)
		if err != nil {
			return nil, err
		}

		if !o {
			res[rk] = cv
		}
	}

	return res, nil
}

// ToAPI is a function that converts filled Terraform user configuration schema to API compatible format.
func ToAPI(st userconfig.SchemaType, n string, d resourceDatable) (map[string]interface{}, error) {
	var res map[string]interface{}

	// fk is a full key slice. We use it to get the full key path to the property in the Terraform user configuration.
	fk := []string{fmt.Sprintf("%s_user_config", n)}

	tp, ok := d.GetOk(fk[0])
	if !ok || tp == nil {
		return res, nil
	}

	tpa, ok := tp.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s (%d): not a slice", n, st)
	}

	ftp := tpa[0]
	if ftp == nil {
		return res, nil
	}

	ftpa, ok := ftp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s.0 (%d): not a map", n, st)
	}

	p, err := props(st, n)
	if err != nil {
		return nil, err
	}

	res, err = propsToAPI(n, append(fk, "0"), ftpa, p, d)
	if err != nil {
		return nil, err
	}

	return res, nil
}
