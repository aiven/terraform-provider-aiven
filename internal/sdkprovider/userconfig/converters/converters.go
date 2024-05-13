// This package provides functions that convert terraform values to JSON and back (from TF to DTO).
// Prerequisites:
// - in terraform a single _object_ is a list with a single element.
//   Hence, `{...}` must be turned into `[{...}]` when value comes _from_ JSON
//   and _unwrapped_ when it goes to JSON
// - terraform doesn't support multiple types per field.
//   Those fields are split into per-type fields: foo_string, foo_object.
//   This means those "virtual" fields must be mapped to a "real" field.
// - A list of objects might change its element order when data is fetched from Aiven.
//   In that case, values _must_ be sorted according to the state values.
//   Otherwise, TF will output a diff.
// - A userconfig values can't be removed on Aiven side.
//   They can be overridden only.

package converters

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/service"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/serviceintegration"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/serviceintegrationendpoint"
)

const (
	userConfigSuffix   = "_user_config"
	AllowIPFilterPurge = "AIVEN_ALLOW_IP_FILTER_PURGE"
)

type userConfigType string

const (
	ServiceUserConfig                    userConfigType = "service"
	ServiceIntegrationUserConfig         userConfigType = "service integration"
	ServiceIntegrationEndpointUserConfig userConfigType = "service integration endpoint"
)

// userConfigKey provides a single source of truth for a field naming
func userConfigKey(kind userConfigType, name string) string {
	switch kind {
	case ServiceIntegrationEndpointUserConfig:
		switch name {
		case "external_google_cloud_bigquery", "external_postgresql":
			// legacy fields
			return name
		}
	}
	return name + userConfigSuffix
}

// getUserConfig returns user config for the given kind and name
func getUserConfig(kind userConfigType, name string) *schema.Schema {
	switch kind {
	case ServiceUserConfig:
		return service.GetUserConfig(name)
	case ServiceIntegrationUserConfig:
		return serviceintegration.GetUserConfig(name)
	case ServiceIntegrationEndpointUserConfig:
		return serviceintegrationendpoint.GetUserConfig(name)
	}
	return nil
}

// getFieldMapping json field for tf field might be different, returns the mapping
func getFieldMapping(kind userConfigType, name string) map[string]string {
	switch kind {
	case ServiceUserConfig:
		return service.GetFieldMapping(name)
	case ServiceIntegrationUserConfig:
		return serviceintegration.GetFieldMapping(name)
	case ServiceIntegrationEndpointUserConfig:
		return serviceintegrationendpoint.GetFieldMapping(name)
	}
	return nil
}

// SetUserConfig sets user config schema for given kind and name
func SetUserConfig(kind userConfigType, name string, s map[string]*schema.Schema) {
	userConfig := getUserConfig(kind, name)
	if userConfig == nil {
		panic(fmt.Sprintf("unknown user config for %s type %q", kind, name))
	}
	s[userConfigKey(kind, name)] = userConfig
}

func Expand(kind userConfigType, name string, d *schema.ResourceData) (map[string]any, error) {
	m, err := expand(kind, name, d)
	if err != nil {
		return nil, fmt.Errorf("error converting user config options for %s type %q to API format: %w", kind, name, err)
	}
	return m, nil
}

// expand expands schema.ResourceData into a DTO map.
// It takes schema.Schema to know how to turn a TF item into json.
func expand(kind userConfigType, name string, d *schema.ResourceData) (map[string]any, error) {
	if getUserConfig(kind, name) == nil {
		// does not have a user config for given kind and name
		return nil, nil
	}

	key := userConfigKey(kind, name)
	state := &stateCompose{
		key:      key,
		path:     key + ".0", // starts from root user config
		schema:   getUserConfig(kind, name),
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

	// Renames ip_filter_object/string to ip_filter
	renameAliasesToDto(kind, name, dto)

	if v, ok := dto["ip_filter"].([]any); ok && len(v) == 0 {
		if _, ok := os.LookupEnv(AllowIPFilterPurge); !ok {
			return nil, fmt.Errorf(
				"ip_filter list is empty, but %[1]s is not set. Please set "+
					"%[1]s to confirm that you want to remove all IP filters, which is going "+
					"to block all traffic to the service",
				AllowIPFilterPurge,
			)
		}
	}
	return dto, nil
}

// stateCompose combines "raw state" and schema.ResourceData
// With the state it is possible to say "if value is null", hence if it is defined by user.
// With schema.ResourceData, you get the value that might be a zero-value.
type stateCompose struct {
	key      string               // state attribute name or schema.ResourceData key
	path     string               // schema.ResourceData path, i.e., foo.0.bar.0.baz to get the value
	schema   *schema.Schema       // tf schema
	config   cty.Value            // tf file state, it knows if resource value is null
	resource *schema.ResourceData // tf resource that has both tf state and data that is received from the API
}

// listItems returns a list of object's states
// Must not use it with scalar types, because "schema" expects to have Resource
func (s *stateCompose) listItems() (result []*stateCompose) {
	if s.config.IsNull() {
		return result
	}

	for i, v := range s.config.AsValueSlice() {
		c := &stateCompose{
			key:      s.key,
			path:     fmt.Sprintf("%s.%d", s.path, i),
			schema:   s.schema, // object is a list with one item, hence the same schema
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
func (s *stateCompose) valueType() schema.ValueType {
	return s.schema.Type
}

func (s *stateCompose) value() any {
	return s.resource.Get(s.path)
}

// isNull returns true if value exist in tf file
func (s *stateCompose) isNull() bool {
	// By some reason iterable values are always not null
	return s.config.IsNull() || s.config.CanIterateElements() && s.config.LengthInt() == 0
}

// hasChange tells if the field has been changed
func (s *stateCompose) hasChange() bool {
	return s.resource.HasChange(s.path)
}

func expandObj(s *stateCompose) (map[string]any, error) {
	m := make(map[string]any)
	for k, v := range s.objectProperties() {
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

func expandScalar(s *stateCompose) (any, error) {
	if s.isNull() {
		// Null scalar, no value in the config
		return nil, nil
	}
	return s.value(), nil
}

// expandAttr returns go value
func expandAttr(s *stateCompose) (any, error) {
	// Scalar values
	switch s.valueType() {
	case schema.TypeSet, schema.TypeList:
		// See below
	case schema.TypeMap:
		return expandMap(s)
	default:
		return expandScalar(s)
	}

	// Here go schema.TypeMap, schema.TypeSet, schema.TypeList
	if s.isNull() && !s.hasChange() {
		// A value that hasn't been sent by user yet.
		// But have been received from the API.
		return nil, nil
	}

	if s.valueType() == schema.TypeSet {
		return expandSet(s)
	}

	return expandList(s)
}

// expandMap must return "any" type to discern "nil" and empty "map"
// to send empty "map" and skip "nil"
func expandMap(s *stateCompose) (any, error) {
	if s.config.IsNull() {
		return nil, nil
	}

	value := s.value()
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%q expected to be map, but got %T", s.path, value)
	}

	// Sends map fields which are in config
	result := make(map[string]any)
	for k, v := range s.config.AsValueMap() {
		if !v.IsNull() {
			result[k] = m[k]
		}
	}
	return result, nil
}

func expandSet(s *stateCompose) ([]any, error) {
	// If the value was removed, sends an empty set
	// Warning: can't handle nested (complex) objects.
	// Use schema.TypeList instead
	result := make([]any, 0)
	if s.config.IsNull() {
		// Makes possible to send ip_filter=[] to clear the remote list.
		return result, nil
	}

	return s.value().(*schema.Set).List(), nil
}

// expandList returns a list of elements or a single object,
// because in TF an object is a list with a single element
func expandList(s *stateCompose) (any, error) {
	// schema.TypeList
	_, isObjList := s.schema.Elem.(*schema.Resource)
	states := s.listItems()
	items := make([]any, 0, len(states))
	for i := range states {
		var exp any
		var err error
		if isObjList {
			exp, err = expandObj(states[i])
		} else {
			exp, err = expandScalar(states[i])
		}

		if err != nil {
			return nil, err
		}

		// If an object is not empty
		if exp != nil {
			items = append(items, exp)
		}
	}

	// If schema.TypeList && MaxItems == 1, then it is an object
	if isObjList && s.schema.MaxItems == 1 {
		switch len(items) {
		case 1:
			// A plain object (in TF a list with one object is an object)
			return items[0], nil
		case 0:
			// The object has no state or removed.
			// We can't remove objects from state, so send a nil.
			return nil, nil
		}
	}

	// A list of scalars
	return items, nil
}

func Flatten(kind userConfigType, name string, d *schema.ResourceData, dto map[string]any) error {
	err := flatten(kind, name, d, dto)
	if err != nil {
		return fmt.Errorf("error converting user config options for %s type %q from API format: %w", kind, name, err)
	}
	return nil
}

// flatten flattens DTO into a terraform compatible object and sets value to the field
func flatten(kind userConfigType, name string, d *schema.ResourceData, dto map[string]any) error {
	if getUserConfig(kind, name) == nil {
		// does not have a user config for given kind and name
		return nil
	}

	key := userConfigKey(kind, name)
	prefix := fmt.Sprintf("%s.0.", key)

	// Renames ip_filter to ip_filter_object
	err := renameAliasesToTfo(kind, name, dto, d)
	if err != nil {
		return err
	}

	// Copies "create only" fields from the original config.
	// Like admin_password, that is received only on POST request when service is created.
	for _, k := range createOnlyFields() {
		v, ok := d.GetOk(prefix + k)
		if ok {
			dto[k] = v
		}
	}

	s := getUserConfig(kind, name)
	r := s.Elem.(*schema.Resource)
	tfo, err := flattenObj(r.Schema, dto)
	if tfo == nil || err != nil {
		return err
	}

	return d.Set(key, []map[string]any{tfo})
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
	case schema.TypeMap:
		return data, nil
	}

	// A set can contain scalars only
	scalar, scalarOk := s.Elem.(*schema.Schema)
	if scalarOk {
		switch s.Type {
		case schema.TypeList:
			return data.([]any), nil
		case schema.TypeSet:
			values := make([]any, 0)
			for _, v := range data.([]any) {
				val, err := flattenAttr(scalar, v)
				if err != nil {
					return nil, err
				}
				values = append(values, val)
			}
			return schema.NewSet(schema.HashSchema(scalar), values), nil
		}
	}

	// Single object or list of objects
	resource := s.Elem.(*schema.Resource)
	var list []any
	if o, isObject := data.(map[string]any); isObject {
		// Single object, but it is a list with one element for terraform
		if len(o) != 0 {
			list = append(list, o)
		}
	} else {
		// List of objects
		list = data.([]any)
	}

	return flattenList(resource.Schema, list)
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

// createOnlyFields these fields are received on POST request only
func createOnlyFields() []string {
	return []string{
		"admin_username",
		"admin_password",
	}
}
