package converters

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const userConfigSuffix = "_user_config"

// Expand expands schema.ResourceData into a DTO map
func Expand(kind string, s *schema.Schema, d *schema.ResourceData) (map[string]any, error) {
	key := kind + userConfigSuffix
	state := &stateCompose{
		key:      key,
		path:     key + ".0", // starts from root user config
		schema:   s,
		resource: d,
	}

	// When "configs" is empty, then we need to delete all arrays in it.
	// That's why it doesn't exit here.
	configs := d.GetRawConfig().GetAttr(key).AsValueSlice()
	if len(configs) > 0 {
		state.config = configs[0]
	}

	dto, err := expandObj(state)
	if err != nil {
		return nil, err
	}

	renameAliases(dto)
	return dto, nil
}

// stateCompose combines "raw state" and schema.ResourceData
// With the state it is possible to say "if value is null", hence is defined by user.
// With schema.ResourceData you get the value.
type stateCompose struct {
	key      string // state attribute name or schema.ResourceData key
	path     string // schema.ResourceData path, i.e. foo.0.bar.0.baz
	schema   *schema.Schema
	config   cty.Value
	resource *schema.ResourceData
}

// setItems returns schema.Set values that has state.
func (s *stateCompose) setItems() []any {
	result := make([]any, 0)
	if s.config.IsNull() {
		// Makes possible to send ip_filter=[] to clear the remote list.
		return result
	}

	// Builds elements hash map
	hashes := make(map[string]bool, s.config.LengthInt())
	for _, item := range s.config.AsValueSlice() {
		if item.Type() == cty.String {
			hashes[item.AsString()] = true
		} else {
			hashes[item.AsBigFloat().String()] = true
		}
	}

	// Picks up values with a state only
	for _, v := range s.get().(*schema.Set).List() {
		if hashes[fmt.Sprintf("%v", v)] {
			result = append(result, v)
		}
	}
	return result
}

// listItems returns a list of object's states
func (s *stateCompose) listItems() (result []*stateCompose) {
	if s.config.IsNull() {
		return result
	}
	for i, v := range s.config.AsValueSlice() {
		c := &stateCompose{
			key:      s.key,
			path:     fmt.Sprintf("%s.%d", s.path, i),
			schema:   s.schema,
			config:   v,
			resource: s.resource,
		}
		result = append(result, c)
	}
	return result
}

// objectProperties returns object's properties states
func (s *stateCompose) objectProperties() map[string]*stateCompose {
	props := make(map[string]*stateCompose)
	res := s.schema.Elem.(*schema.Resource)
	for key, subSchema := range res.Schema {
		if subSchema.ForceNew && !s.resource.IsNewResource() {
			continue
		}

		var config cty.Value
		if !s.config.IsNull() {
			// Can't get value from nil
			config = s.config.GetAttr(key)
		}

		p := &stateCompose{
			key:      key,
			path:     fmt.Sprintf("%s.%s", s.path, key),
			resource: s.resource,
			config:   config,
			schema:   subSchema,
		}

		props[key] = p
	}
	return props
}
func (s *stateCompose) get() any {
	return s.resource.Get(s.path)
}

func (s *stateCompose) isNull() bool {
	return s.config.IsNull()
}

func (s *stateCompose) hasChange() bool {
	return s.resource.HasChange(s.path)
}

func expandObj(state *stateCompose) (map[string]any, error) {
	m := make(map[string]any)
	for k, v := range state.objectProperties() {
		value, err := expandAttr(v)
		if err != nil {
			return nil, fmt.Errorf("%q field conversion error: %w", k, err)
		}
		if value != nil {
			m[k] = value
		}
	}
	return m, nil
}

// expandAttr returns go value
func expandAttr(state *stateCompose) (any, error) {
	switch state.schema.Type {
	case schema.TypeString, schema.TypeBool, schema.TypeInt, schema.TypeFloat:
		if state.isNull() {
			// Null scalar, no value in the config
			return nil, nil
		}
		return state.get(), nil
	}

	if state.schema.Type == schema.TypeSet {
		if state.isNull() && !state.hasChange() {
			// A value that hasn't been sent by user yet.
			// But has been received from the API.
			return nil, nil
		}

		return state.setItems(), nil
	}

	// schema.TypeList
	states := state.listItems()
	items := make([]any, 0, len(states))
	for i := range states {
		exp, err := expandObj(states[i])
		if err != nil {
			return nil, err
		}
		// If an object is not empty
		if exp != nil {
			items = append(items, exp)
		}
	}

	// That is a list of objects
	if state.schema.MaxItems != 1 {
		return items, nil
	}

	// If schema.TypeList && MaxItems == 1, then it is an object
	switch len(items) {
	case 1:
		// A plain object (in TF a list with one object is an object)
		return items[0], nil
	case 0:
		// The object has no state or removed.
		// We can't remove objects from state, so send a nil.
		return nil, nil
	default:
		// If MaxItems == 1, then this shouldn't ever happen
		return nil, fmt.Errorf("unexpected list length %d for key %s", len(items), state.key)
	}
}

func renameAliases(dto map[string]any) {
	keys := []struct {
		path    string
		name    string
		aliases []string
	}{
		{
			path:    "", // root
			name:    "ip_filter",
			aliases: []string{"ip_filter_string", "ip_filter_object"},
		},
		{
			path:    "rules.0.mapping",
			name:    "namespaces",
			aliases: []string{"namespaces_string", "namespaces_object"},
		},
	}

	for _, key := range keys {
		var branches []map[string]any
		if key.path == "" {
			branches = append(branches, dto)
		} else {
			v, ok := drillKey(dto, key.path)
			if !ok {
				// branch does not exist, nothing to do
				continue
			}

			// It can be a list of maps, or just one map
			switch v.(type) {
			case []any:
				branches = asMapList(v)
			default:
				branches = append(branches, v.(map[string]any))
			}
		}

		for _, branch := range branches {
			for _, alias := range key.aliases {
				// Copies only non-zero values.
				// For instance: foo=[], foo_string=[val], foo_object=[]
				// "foo_object" shouldn't override "foo_string"
				if v, ok := branch[alias]; ok {
					// It is valid to send an empty list.
					// So we must choose non-empty alias.
					if a, ok := v.([]any); ok && len(a) > 0 {
						branch[key.name] = v
					}
					delete(branch, alias)
				}
			}
		}
	}
}

// Flatten flattens DTO into a terraform compatible object
func Flatten(kind string, s *schema.Schema, d *schema.ResourceData, dto map[string]any) ([]map[string]any, error) {
	withPrefix := func(v string) string {
		return fmt.Sprintf("%s%s.0.%s", kind, userConfigSuffix, v)
	}

	// Renames ip_filter field
	if _, ok := dto["ip_filter"]; ok {
		assignAlias(d, withPrefix("ip_filter"), dto, "ip_filter", "network")
	}

	// Renames namespaces field
	if mapping, ok := drillKey(dto, "rules.0.mapping"); ok {
		assignAlias(d, withPrefix("rules.0.mapping.0.namespaces"), mapping.(map[string]any), "namespaces", "name")
	}

	// Copies "create only" fields from the original config.
	// Like admin_password, that is received only on POST request when service is created.
	for _, k := range createOnlyFields() {
		v, ok := d.GetOk(withPrefix(k))
		if ok {
			dto[k] = v
		}
	}

	r := s.Elem.(*schema.Resource)
	tfo, err := flattenObj(r.Schema, dto)
	if tfo == nil || err != nil {
		return nil, err
	}
	return []map[string]any{tfo}, nil
}

func flattenObj(s map[string]*schema.Schema, dto map[string]any) (map[string]any, error) {
	tfo := make(map[string]any)
	for k, v := range s {
		vv, ok := dto[k]
		if !ok {
			continue
		}

		if vv == nil {
			continue
		}

		value, err := flattenAttr(v, vv)
		if err != nil {
			return nil, fmt.Errorf("%q field conversion error: %w", k, err)
		}

		if value != nil {
			tfo[k] = value
		}
	}
	if len(tfo) == 0 {
		return nil, nil
	}
	return tfo, nil
}

func flattenAttr(s *schema.Schema, data any) (any, error) {
	switch s.Type {
	case schema.TypeString:
		return castType[string](data)
	case schema.TypeBool:
		return castType[bool](data)
	case schema.TypeInt:
		i, err := data.(json.Number).Int64()
		return int(i), err
	case schema.TypeFloat:
		return data.(json.Number).Float64()
	}

	scalarSchema, isScalar := s.Elem.(*schema.Schema)
	if isScalar {
		values := make([]any, 0)
		for _, v := range data.([]any) {
			val, err := flattenAttr(scalarSchema, v)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		return schema.NewSet(schema.HashSchema(scalarSchema), values), nil
	}

	// Single object or list of objects
	r := s.Elem.(*schema.Resource)
	if s.Type == schema.TypeList {
		var list []any
		if o, isObject := data.(map[string]any); isObject {
			// Single object
			if len(o) != 0 {
				list = append(list, o)
			}
		} else {
			// List of objects
			list = data.([]any)
		}

		return flattenList(r.Schema, list)
	}

	// Array of scalars
	items, err := flattenList(r.Schema, data.([]any))
	if items == nil || err != nil {
		return nil, err
	}

	return schema.NewSet(schema.HashResource(r), items), nil
}

func flattenList(s map[string]*schema.Schema, list []any) ([]any, error) {
	if len(list) == 0 {
		return nil, nil
	}

	items := make([]any, 0, len(list))
	for _, item := range list {
		v, err := flattenObj(s, item.(map[string]any))
		if err != nil {
			return nil, err
		}
		if v != nil {
			items = append(items, v)
		}
	}
	return items, nil
}

// assignAlias renames keys for multi-typed properties, i.e. ip_filter -> [ip_filter_string, ip_filter_object]
func assignAlias(d *schema.ResourceData, path string, dto map[string]any, key, sortBy string) {
	values, ok := dto[key].([]any)
	if !ok || len(values) == 0 {
		return
	}

	var suffix string
	const (
		str = "_string"
		obj = "_object"
	)

	// If DTO has objects, then it is foo_object
	if _, ok := values[0].(map[string]any); ok {
		suffix = obj

		// State objects have specific order.
		// Must sort DTO objects, otherwise diff shows changes.
		if inStateObjs, ok := d.GetOk(path + obj); ok {
			dto[key] = sortByKey(sortBy, inStateObjs, dto[key])
		}
	}

	// If the state has foo_string, the user has new key
	if _, ok := d.GetOk(path + str); ok {
		suffix = str
	}

	if suffix != "" {
		dto[key+suffix] = dto[key]
		delete(dto, key)
	}
}

// createOnlyFields these fields are received on POST request only
func createOnlyFields() []string {
	return []string{
		"admin_username",
		"admin_password",
	}
}
