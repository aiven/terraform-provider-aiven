package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"
)

type ResourceData interface {
	Get(key string) any
	GetOk(key string) (any, bool)
	GetState(key string) any
	HasChange(key string) bool
	Set(key string, value any) error
	SetID(parts ...string) error
	ID() string
	Expand(out any, modifiers ...MapModifier) error
	Flatten(in any, modifiers ...MapModifier) error
	tfValue() tftypes.Value
}

type resourceData struct {
	schema              *Schema
	plan, state, config map[string]any
	idFields            []string
}

var _ ResourceData = (*resourceData)(nil)

// ID returns the value of the "id" field
// which is mandatory for all resources and datasources.
func (d *resourceData) ID() string {
	return d.Get("id").(string)
}

// SetID sets the value of the "id" field in path-like format.
func (d *resourceData) SetID(parts ...string) error {
	return d.Set("id", strings.Join(parts, "/"))
}

// currentState returns the current state map.
// Depending on the operation values are taken and written back to the same map.
//
// Create operation reads plan and config, writes back to the plan
// Update operation reads plan, state and config, writes back to the plan
// Read and Delete operations read state, writes back to the state
func (d *resourceData) currentState() map[string]any {
	if d.plan != nil {
		return d.plan
	}
	if d.state != nil {
		return d.state
	}
	return d.config
}

// Get returns value from plan, state or config that can be safely cast to the expected type.
func (d *resourceData) Get(key string) any {
	v, _ := d.GetOk(key)
	return v
}

// GetOk returns a value by a path.
// Tries to get the value from the plan, then from the state, then from the config.
// Returns false if the value is not found.
func (d *resourceData) GetOk(key string) (any, bool) {
	v, sch, ok, err := getOk(d.schema, d.plan, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for %q from plan: %w", key, err))
	}

	if ok {
		return v, ok
	}

	if sch.Computed || d.plan == nil {
		// 1. Computed fields can be read from the state
		// 2. Read and Delete operations do not have "plan", everything is read from the state
		v, _, ok, err = getOk(d.schema, d.state, key)
		if err != nil {
			panic(fmt.Errorf("failed to get value for %q from state: %w", key, err))
		}

		if ok {
			return v, ok
		}
	}

	// 1. Datasource reads config only
	// 2. WriteOnly fields are read from the config
	v, _, ok, err = getOk(d.schema, d.config, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for %q from config: %w", key, err))
	}
	return v, ok
}

// GetState returns a value from the state.
// Some resources may change ID fields during the Update operation.
// This will place new values into the URL, and the resource may not be found on the backend.
func (d *resourceData) GetState(key string) any {
	if d.state == nil {
		panic(fmt.Errorf("state is nil"))
	}
	v, _, _, err := getOk(d.schema, d.state, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for %q from state: %w", key, err))
	}
	return v
}

func (d *resourceData) HasChange(key string) bool {
	if (d.plan == nil) && (d.state == nil) {
		return false
	}

	planVal, _, pOK, err := getOk(d.schema, d.plan, key)
	if err != nil {
		panic(fmt.Errorf("failed to get plan value for %q: %w", key, err))
	}

	stateVal, _, sOK, err := getOk(d.schema, d.state, key)
	if err != nil {
		panic(fmt.Errorf("failed to get state value for %q: %w", key, err))
	}
	if pOK != sOK {
		return true
	}
	return !cmp.Equal(planVal, stateVal)
}

// Set sets a value by a path.
// Returns an error if the key is not found in the schema.
func (d *resourceData) Set(key string, value any) error {
	return d.set(key, value, false)
}

func (d *resourceData) set(key string, value any, ignoreUnknown bool) error {
	prop, ok := d.schema.Properties[key]
	if !ok {
		return fmt.Errorf("key %q not found in schema", key)
	}

	state := d.currentState()
	if value == nil {
		// Nil is a special case, means "remove"
		delete(state, key)
	} else {
		norm, err := normalizeAny(prop, value, ignoreUnknown, true)
		if err != nil {
			return err
		}
		state[key] = norm
	}
	return nil
}

// Expand converts the plan to Request.
func (d *resourceData) Expand(out any, modifiers ...MapModifier) error {
	if d.plan == nil {
		return fmt.Errorf("no plan provided")
	}

	var m map[string]any
	err := remarshal(&d.plan, &m)
	if err != nil {
		return err
	}

	norm, err := normalizeTyped(d.schema, m, false, false)
	if err != nil {
		return err
	}

	// Sets empty strings and arrays for all removed values to override backend data.
	for k := range d.state {
		sch, ok := d.schema.Properties[k]
		if !ok || sch.Computed {
			// 1. !ok: this is unusual, the field is not in the schema.
			//    Not sure if this is possible.
			// 2. sch.Computed — computed fields when removed just preserve the value.
			//    We don't need to override/delete them on the backend.
			continue
		}

		// If the string/array/set field was removed, set it to empty value
		// to override the backend value.
		_, ok = norm[k]
		if !ok {
			switch sch.Type {
			case SchemaTypeString, SchemaTypeList, SchemaTypeSet:
				norm[k] = zeroValue(sch.Type)
			}
		}
	}

	for _, modifier := range modifiers {
		if err := modifier(d, norm); err != nil {
			return err
		}
	}
	return remarshal(&norm, out)
}

// Flatten converts the Response to State.
func (d *resourceData) Flatten(in any, modifiers ...MapModifier) error {
	var m map[string]any
	err := remarshal(in, &m)
	if err != nil {
		return err
	}

	for _, modifier := range modifiers {
		if err := modifier(d, m); err != nil {
			return err
		}
	}

	norm, err := normalizeTyped(d.schema, m, true, true)
	if err != nil {
		return err
	}

	// todo: remove stale data
	state := d.currentState()
	for k, v := range norm {
		state[k] = v
	}

	id := make([]string, len(d.idFields))
	for i, name := range d.idFields {
		v, ok := d.GetOk(name)
		if !ok {
			return fmt.Errorf("no value for id field %q", name)
		}
		id[i] = v.(string)
	}

	return d.SetID(id...)
}

// tfValue converts the current state to tftypes.Value to write it to the user's state.
func (d *resourceData) tfValue() tftypes.Value {
	v, err := toTFValue(d.schema, d.currentState())
	if err != nil {
		panic(fmt.Errorf("failed to convert state to tftypes.Value: %w", err))
	}
	return v
}

// getOk returns a value by a path:
//
//	{
//	  "foo": [{"bar": [{"baz": "wrong!"}, {"baz": "here!"}]}]
//	}
//
// Returns "here!" when path is "foo.0.bar.1.baz"
//
// - When the path is valid (exists in the schema), but the value is not found, returns zero value.
// - When the path is invalid, returns an error.
// - A set can't be indexed, returns an error.
func getOk(sch *Schema, data any, path string) (any, *Schema, bool, error) {
	if path == "" {
		return nil, nil, false, fmt.Errorf("key is empty")
	}

	if sch == nil {
		return nil, nil, false, fmt.Errorf("schema is nil")
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		switch sch.Type {
		case SchemaTypeSet:
			return nil, nil, false, fmt.Errorf("invalid path %q: set is not supported", path)
		case SchemaTypeList:
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, nil, false, fmt.Errorf("invalid index %q for %q: %w", part, path, err)
			}
			sch = sch.Items
			list, ok := data.([]any)
			if !ok {
				return nil, nil, false, fmt.Errorf("expected list %q at %q, got %T", part, path, data)
			}

			// Handle empty lists and index overflows gracefully.
			if index < 0 || index >= len(list) {
				return nil, nil, false, fmt.Errorf("invalid list %q at %q: index %d out of range (len=%d)", part, path, index, len(list))
			}

			data = list[index]
		case SchemaTypeObject:
			object, ok := data.(map[string]any)
			if !ok {
				return nil, nil, false, fmt.Errorf("expected object %q at %q, got %T", part, path, data)
			}

			sch, ok = sch.Properties[part]
			if !ok {
				return nil, nil, false, fmt.Errorf("key %q at path %q not found in schema", part, path)
			}

			prop, ok := object[part]
			if !ok {
				return zeroValue(sch.Type), sch, false, nil
			}
			data = prop
		case SchemaTypeMap:
			dict, ok := data.(map[string]any)
			if !ok {
				return nil, nil, false, fmt.Errorf("expected map %q at %q, got %T", part, path, data)
			}

			sch = sch.Items
			val, ok := dict[part]
			if !ok {
				return zeroValue(sch.Type), sch, false, nil
			}
			data = val
		case SchemaTypeString, SchemaTypeFloat, SchemaTypeInt, SchemaTypeBool:
			val, err := normalizeAny(sch, data, false, false)
			if err != nil {
				return nil, nil, false, fmt.Errorf("invalid value for %q at %q: %w", part, path, err)
			}
			data = val
		}
	}

	return data, sch, true, nil
}

func NewResourceDataFromMaps(schema *Schema, idFields []string, plan, state, config map[string]any) (ResourceData, error) {
	rd := &resourceData{
		schema:   schema,
		idFields: idFields,
	}
	rd.plan = plan
	rd.state = state
	rd.config = config
	return rd, nil
}

func NewResourceData(schema *Schema, idFields []string, plan *tfsdk.Plan, state *tfsdk.State, config *tfsdk.Config) (ResourceData, error) {
	var err error
	var planMap, stateMap, configMap map[string]any
	if plan == nil && state == nil && config == nil {
		return nil, fmt.Errorf("plan, state and config are nil")
	}

	if plan != nil && !plan.Raw.IsNull() {
		planMap, err = fromTFValue(schema, plan.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode plan: %w", err)
		}
	}

	if state != nil && !state.Raw.IsNull() {
		stateMap, err = fromTFValue(schema, state.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode state: %w", err)
		}
	}

	if config != nil && !config.Raw.IsNull() {
		configMap, err = fromTFValue(schema, config.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config: %w", err)
		}
	}

	return NewResourceDataFromMaps(schema, idFields, planMap, stateMap, configMap)
}

// normalizeTyped see normalizeAny for more details.
func normalizeTyped[T any](sch *Schema, m T, ignoreUnknown bool, setFlow bool) (T, error) {
	v, err := normalizeAny(sch, m, ignoreUnknown, setFlow)
	if err != nil {
		return m, err
	}

	t, ok := v.(T)
	if !ok {
		return t, fmt.Errorf("expected %T, got %T", t, m)
	}
	return t, nil
}

// dereference dereferences a pointer: only a single level of indirection is allowed.
func dereference(value any) (any, error) {
	for i := 0; ; i++ {
		if lo.IsNil(value) {
			return nil, nil
		}

		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			value = val.Elem().Interface()
		} else {
			break
		}

		if i == 1 {
			return nil, fmt.Errorf("pointer to pointer not allowed")
		}
	}
	return value, nil
}

// normalizeAny
// For reads: returns a typed value from an "any" type, so you can safely do "any.(int)".
// For writes: validates and casts the value according to the schema.
// ignoreUnknown — when true, unknown fields are ignored (useful when ResourceData.Flatten sets values; new fields from the API can be safely skipped).
// setFlow — when true, objects must be wrapped in a list. See Schema.IsObject for more details.
func normalizeAny(sch *Schema, value any, ignoreUnknown bool, setFlow bool) (any, error) {
	value, err := dereference(value)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return zeroValue(sch.Type), nil
	}

	if isScalar(sch.Type) {
		str := fmt.Sprint(value)
		switch sch.Type {
		case SchemaTypeString:
			if str == "" && setFlow {
				// We don't want to store empty strings in the state
				return nil, nil
			}
			return str, nil
		case SchemaTypeInt:
			return strconv.ParseInt(str, 10, 64)
		case SchemaTypeFloat:
			return strconv.ParseFloat(str, 64)
		case SchemaTypeBool:
			return strconv.ParseBool(str)
		}
	}

	switch sch.Type {
	case SchemaTypeSet, SchemaTypeList:
		list, ok := value.([]any)
		if !ok {
			if sch.IsObject && setFlow {
				// Objects are lists of a single object.
				list = append(list, value)
			} else {
				return nil, fmt.Errorf("expected set or list, got %T", value)
			}
		}
		norm := make([]any, 0, len(list))
		for _, elem := range list {
			item, err := normalizeAny(sch.Items, elem, ignoreUnknown, setFlow)
			if err != nil {
				return nil, err
			}
			if sch.IsObject && !setFlow {
				// Objects are lists of a single object.
				// We need to return the object directly, not the list.
				return item, nil
			}
			norm = append(norm, item)
		}
		return norm, nil
	case SchemaTypeMap:
		dict, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", value)
		}
		norm := make(map[string]any, len(dict))
		for key, elem := range dict {
			item, err := normalizeAny(sch.Items, elem, ignoreUnknown, setFlow)
			if err != nil {
				return nil, err
			}
			norm[key] = item
		}
		return norm, nil
	case SchemaTypeObject:
		dict, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object, got %T", value)
		}
		norm := make(map[string]any, len(dict))
		for key, elem := range dict {
			prop, ok := sch.Properties[key]
			if !ok {
				if ignoreUnknown {
					// This is likely a new field from the API or a field we do not want to expose to the user.
					continue
				}
				return nil, fmt.Errorf("unknown property %q", key)
			}
			item, err := normalizeAny(prop, elem, ignoreUnknown, setFlow)
			if err != nil {
				return nil, err
			}
			norm[key] = item
		}
		return norm, nil
	default:
		return nil, fmt.Errorf("can't normalize type: %T", value)
	}
}

func zeroValue(kind SchemaType) any {
	switch kind {
	case SchemaTypeString:
		return ""
	case SchemaTypeInt:
		return 0
	case SchemaTypeFloat:
		return 0.0
	case SchemaTypeBool:
		return false
	case SchemaTypeList, SchemaTypeSet:
		return []any{}
	case SchemaTypeMap, SchemaTypeObject:
		return map[string]any{}
	}
	panic(fmt.Sprintf("unknown schema type: %q", kind))
}

func remarshal(in, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	// Uses json.Number to avoid int->float64->int overflow issue.
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	return d.Decode(out)
}
