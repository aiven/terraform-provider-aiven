package apiconvert

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// resourceDatable is an interface that allows to get the resource data from the schema.
// This is needed to be able to test the conversion functions. See schema.ResourceData for more.
type resourceDatable interface {
	GetOk(string) (interface{}, bool)
	HasChange(string) bool
	IsNewResource() bool
}

var (
	// dotAnyDotNumberRegExp is a regular expression that matches a string that matches:
	//   1. key.1.key2.0.key3.2.key.5
	//   2. key123.0
	//   3. key.1
	//   4. key2.9
	//   5. key..8
	// and does not match:
	//   1. key.key2
	//   2. key.01
	//   3. key.abc
	//   4. .1
	//   5. key.
	dotAnyDotNumberRegExp = regexp.MustCompile(`.+\.[0-9]$`)

	// dotNumberEOLOrDotRegExp is a regular expression that matches a string that matches:
	//   1. .5 (match: .5)
	//   2. .9. (match: .9.)
	//   3. 0.1 (match: .1)
	//   4. key.2 (match: .2)
	//   5. 1.2.3 (match: .2.)
	//   6. key..8 (match: .8)
	// and does not match:
	//   1. .123
	//   2. 1.
	//   3. 1..
	//   4. .5a
	dotNumberEOLOrDotRegExp = regexp.MustCompile(`\.\d($|\.)`)
)

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

		vnc, o, err := itemToAPI(
			n,
			iit,
			append(fk, fmt.Sprintf("%d", i)),
			fmt.Sprintf("%s.%d", k, i),
			vn,
			ii,
			false,
			d,
		)
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

	// Object with only "null" fields becomes nil
	// Which can't be cast into a map
	if fv == nil {
		return res, true, nil
	}

	fva, ok := fv.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%s: not a map", fks)
	}

	ip, ok := i["properties"].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%s (item): properties key not found", fks)
	}

	reqs := map[string]struct{}{}

	if sreqs, ok := i["required"].([]interface{}); ok {
		reqs = userconfig.SliceToKeyedMap(sreqs)
	}

	if !dotAnyDotNumberRegExp.MatchString(fks) {
		fk = append(fk, "0")
	}

	res, err := propsToAPI(n, fk, fva, ip, reqs, d)
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
	ireq bool,
	d resourceDatable,
) (interface{}, bool, error) {
	res := v

	fks := strings.Join(fk, ".")

	// We omit the value if it has no changes in the Terraform user configuration.
	o := !d.HasChange(fks)

	// We need to make sure that if there were any changes to the parent object, we also send the value, even if it
	// was not changed.
	//
	// We check that there are more than three elements in the fk slice, because we don't want to send the value if
	// the parent object is the root object.
	if o && len(fk) > 3 {
		// We find the last index of the dot with a number after it, because we want to check if the parent object
		// was changed.
		match := dotNumberEOLOrDotRegExp.FindAllStringIndex(fks, -1)
		if match != nil {
			// We check if fks exists, i.e. it was set by the user, because if it was not set, we don't want to send
			// the value.
			_, e := d.GetOk(fks)

			// Since Terraform thinks that new array elements are added without "existing", we also send the value if
			// it does not exist, but is not empty either.
			if (e || !reflect.ValueOf(v).IsZero()) && d.HasChange(fks[:match[len(match)-1][0]]) {
				o = false
			}
		}
	}

	// We need to make sure that if the value is required, we send it, even if it has no changes in the Terraform.
	if o && ireq {
		o = false
	}

	// Assert the type of the value to match.
	switch t {
	case "boolean":
		if _, ok := v.(bool); !ok {
			return nil, false, fmt.Errorf("%s: not a boolean", fks)
		}
	case "integer":
		if _, ok := v.(int); !ok {
			return nil, false, fmt.Errorf("%s: not an integer", fks)
		}
	case "number":
		if _, ok := v.(float64); !ok {
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
			// This can be TypeSet
			set, ok := v.(*schema.Set)
			if !ok {
				return nil, false, fmt.Errorf("%s: not slice or set", fks)
			}
			va = set.List()
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

// processManyToOneKeys is a function that processes many to one keys.
// It processes the provided result with keys in their flattened form and sets the many to one key to the value of the
// first flattened key that is not empty, and uses it to send the value to the API.
func processManyToOneKeys(res map[string]interface{}) {
	// mto is a map of many to one keys that exist in the provided properties.
	mto := make(map[string][]string)

	// TODO: Remove all ip_filter and namespaces special cases when these fields are removed.
	for k, v := range res {
		// If the value is a map, we process it recursively.
		if va, ok := v.(map[string]interface{}); ok {
			processManyToOneKeys(va)
		}

		// We ignore untyped keys, because they cannot be many to one.
		if !userconfig.IsKeyTyped(k) && k != "ip_filter" && k != "namespaces" {
			continue
		}

		// rk is the real key, i.e. the key without the suffix.
		rk := k

		if k != "ip_filter" && k != "namespaces" {
			rk = k[:strings.LastIndexByte(k, '_')]
		}

		// If the key does not exist in the map, we create it.
		if _, ok := mto[rk]; !ok {
			mto[rk] = []string{}
		}

		// We append the key to the list of keys that are many to one.
		mto[rk] = append(mto[rk], k)
	}

	// At this point mto looks like this, or similar:
	// map[string][]string{
	//  // ip_filter has two keys set in the user configuration, so we use the first one that is not empty,
	//  // e.g. when user switches from ip_filter to ip_filter_object, we use ip_filter_object.
	// 	"ip_filter": []string{"ip_filter", "ip_filter_object"},
	//  // namespaces has only one key set in the user configuration, so we use it.
	// 	"namespaces": []string{"namespaces"},
	// }

	// We iterate over the map of many to one keys and process them.
	for k, v := range mto {
		// nv is the new value of the key.
		var nv interface{}

		for _, vn := range v {
			// If the many to one key is not set or is empty, we skip it by removing it from the map.
			if rv, ok := res[vn].([]interface{}); ok && len(rv) > 0 {
				nv = rv

				delete(res, vn)
			}
		}

		// Finally, we set the new value of the key.
		res[k] = nv
	}
}

// propsToAPI is a function that converts properties of Terraform user configuration schema to API compatible format.
func propsToAPI(
	n string,
	fk []string,
	tp map[string]interface{},
	p map[string]interface{},
	reqs map[string]struct{},
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

		_, ireq := reqs[k]

		t := ats[0]

		cv, o, err := itemToAPI(n, t, append(fk, k), k, v, ia, ireq, d)
		if err != nil {
			return nil, err
		}

		if !o {
			res[k] = cv
		}
	}

	processManyToOneKeys(res)

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

	p, reqs, err := propsReqs(st, n)
	if err != nil {
		return nil, err
	}

	res, err = propsToAPI(n, append(fk, "0"), ftpa, p, reqs, d)
	if err != nil {
		return nil, err
	}

	return res, nil
}
