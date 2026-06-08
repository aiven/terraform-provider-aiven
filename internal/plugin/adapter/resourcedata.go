package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const idField = "id"

type ResourceData interface {
	Get(key string) any
	GetOk(key string) (any, bool)
	GetState(key string) any
	HasChange(key string) bool
	Set(key string, value any) error
	SetID(parts ...string) error
	ID() string
	IsNewResource() bool
	IsDataSource() bool
	IsResource() bool
	Schema() *Schema
	Expand(out any, modifiers ...MapModifier) error
	Flatten(in any, modifiers ...MapModifier) error
	tfValue() tftypes.Value
}

type resourceData struct {
	schema              *Schema
	plan, state, config map[string]any
	idFields            []string
	isDataSource        bool
	preservePlanValues  bool
}

var _ ResourceData = (*resourceData)(nil)

// ID returns the value of the "id" field
// which is mandatory for all resources and datasources.
func (d *resourceData) ID() string {
	return d.Get(idField).(string)
}

// IsNewResource returns true if the resource is being created.
func (d *resourceData) IsNewResource() bool {
	return d.ID() == ""
}

// IsDataSource returns true if the ResourceData was created for a data source.
func (d *resourceData) IsDataSource() bool {
	return d.isDataSource
}

// IsResource returns true if the ResourceData was created for a managed resource.
func (d *resourceData) IsResource() bool {
	return !d.isDataSource
}

// SetID sets the value of the "id" field in path-like format.
func (d *resourceData) SetID(parts ...string) error {
	return d.Set(idField, strings.Join(parts, "/"))
}

// Schema returns the schema of the resource.
// NOTE: do not modify the schema.
func (d *resourceData) Schema() *Schema {
	return d.schema
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

// GetOk looks up a value by key, checking plan, then config, then state. Returns the value and whether it was found.
// ID fields are read from the state even if the plan is not present.
func (d *resourceData) GetOk(key string) (any, bool) {
	// getOk doesn't fail if plan is nil.
	// It returns a typed zero value.
	v, _, ok, err := getOk(d.schema, d.plan, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for %q from plan: %w", key, err))
	}

	if ok {
		return v, true
	}

	// 1. Datasource reads config only
	// 2. WriteOnly fields are read from the config
	if d.config != nil {
		v, _, ok, err = getOk(d.schema, d.config, key)
		if err != nil {
			panic(fmt.Errorf("failed to get value for %q from config: %w", key, err))
		}

		if ok {
			return v, true
		}
	}

	// State is consulted only in these cases:
	//   - Read/Delete operations (no plan): state is the source of truth.
	//   - "id" and its components: always needed for API calls, even during Update,
	//     where reading other fields from state could resurrect values the user removed.
	isIDComponent := slices.Contains(d.idFields, key)
	if d.state == nil || (d.plan != nil && key != idField && !isIDComponent) {
		return v, ok
	}

	v, _, ok, err = getOk(d.schema, d.state, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for %q from state: %w", key, err))
	}
	if ok {
		return v, true
	}

	// An id-field component may not be stored as its own field in state
	// (e.g. some legacy resources only persist the composite "id"). Try to derive it
	// by splitting the composite id, e.g. "project/vpc_id" -> ["project", "vpc_id"].
	if isIDComponent {
		if idStr, _ := d.state[idField].(string); idStr != "" {
			parts, err := schemautil.SplitResourceID(idStr, len(d.idFields))
			if err != nil {
				panic(fmt.Errorf("failed to split id from state: %w", err))
			}
			return parts[slices.Index(d.idFields, key)], true
		}
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

func (d *resourceData) set(key string, value any, ignoreUnknownKeys bool) error {
	prop, ok := d.schema.Properties[key]
	if !ok {
		return fmt.Errorf("key %q not found in schema", key)
	}

	state := d.currentState()
	if value == nil {
		// Nil is a special case, means "remove"
		delete(state, key)
	} else {
		norm, err := normalizeAny(prop, value, ignoreUnknownKeys, true)
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

	norm, err := normalizeTyped(d.schema, d.plan, false, false)
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
		if !ok && !sch.ZeroNotAllowed {
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

	// The Plugin Framework does not support computed blocks:
	// https://discuss.hashicorp.com/t/provider-plugin-framework-computed-blocks-block-count-changed-from-x-to-y/72955/2
	// Remove computed blocks from resources (data sources are essentially computed) when values are not set.
	// Currently, this is applied only at the top level.
	// todo: remove in v5.0.0, this is a legacy.
	if d.IsResource() && d.config != nil {
		for k, v := range d.schema.Properties {
			if v.IsObject && v.Computed && isEmpty(d.config[k]) {
				delete(norm, k)
			}
		}
	}

	// todo: remove stale data
	state := d.currentState()
	maps.Copy(state, norm)

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

	parts := strings.SplitSeq(path, ".")
	for part := range parts {
		switch sch.Type {
		case SchemaTypeSet:
			return nil, nil, false, fmt.Errorf("invalid path %q: set indexing is not supported", path)
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

		// This is a preserved nil/unknown value for ModifyPlan.
		if _, ok := data.(tftypes.Value); ok {
			return zeroValue(sch.Type), sch, false, nil
		}
	}

	return data, sch, true, nil
}

// ResourceDataOpt configures a resourceData instance during construction.
type ResourceDataOpt func(*resourceData) error

// NewResourceData creates a new ResourceData from the given options.
// Create operation needs WithPlan and WithConfig.
// Update operation needs WithPlan, WithState and WithConfig.
// Read and Delete operations need WithState.
func NewResourceData(schema *Schema, idFields []string, opts ...ResourceDataOpt) (ResourceData, error) {
	rd := &resourceData{
		schema:   schema,
		idFields: idFields,
	}
	for _, opt := range opts {
		if err := opt(rd); err != nil {
			return nil, err
		}
	}
	return rd, nil
}

// WithPreservePlanValues keeps nil/unknown tftypes.Value entries in the plan map.
// Must be applied before WithPlan.
func WithPreservePlanValues() ResourceDataOpt {
	return func(d *resourceData) error {
		d.preservePlanValues = true
		return nil
	}
}

// WithPlan decodes the Terraform plan into ResourceData.
func WithPlan(plan tfsdk.Plan) ResourceDataOpt {
	return func(d *resourceData) error {
		planMap, err := fromTFValue(d.schema, plan.Raw, d.preservePlanValues)
		if err != nil {
			return fmt.Errorf("failed to decode plan: %w", err)
		}
		d.plan = planMap
		return nil
	}
}

// WithState decodes the Terraform state into ResourceData.
func WithState(state tfsdk.State) ResourceDataOpt {
	return func(d *resourceData) error {
		stateMap, err := fromTFValue(d.schema, state.Raw, false)
		if err != nil {
			return fmt.Errorf("failed to decode state: %w", err)
		}
		d.state = stateMap
		return nil
	}
}

// WithConfig decodes the Terraform config into ResourceData.
func WithConfig(config tfsdk.Config) ResourceDataOpt {
	return func(d *resourceData) error {
		configMap, err := fromTFValue(d.schema, config.Raw, false)
		if err != nil {
			return fmt.Errorf("failed to decode config: %w", err)
		}
		d.config = configMap
		return nil
	}
}

// WithIsDataSource marks ResourceData as belonging to a data source.
func WithIsDataSource() ResourceDataOpt {
	return func(d *resourceData) error {
		d.isDataSource = true
		return nil
	}
}

// WithTestPlan sets the plan from a map. For tests only.
func WithTestPlan(plan map[string]any) ResourceDataOpt {
	return func(d *resourceData) error {
		d.plan = plan
		return nil
	}
}

// WithTestState sets the state from a map. For tests only.
func WithTestState(state map[string]any) ResourceDataOpt {
	return func(d *resourceData) error {
		d.state = state
		return nil
	}
}

// WithTestConfig sets the config from a map. For tests only.
func WithTestConfig(config map[string]any) ResourceDataOpt {
	return func(d *resourceData) error {
		d.config = config
		return nil
	}
}

// normalizeTyped see normalizeAny for more details.
func normalizeTyped[T any](sch *Schema, m T, ignoreUnknownKeys bool, setFlow bool) (T, error) {
	v, err := normalizeAny(sch, m, ignoreUnknownKeys, setFlow)
	if err != nil {
		return m, err
	}

	t, ok := v.(T)
	if !ok {
		return t, fmt.Errorf("expected %T, got %T", t, m)
	}
	return t, nil
}

// asAnyMap returns the value as map[string]any. The fast path is a direct
// type assertion; the slow path uses reflection to convert any string-keyed
// map (e.g. map[string]string) so callers don't have to JSON round-trip
// strictly-typed maps to satisfy a map[string]any assertion. Entry values
// are returned as-is; downstream normalization handles their conversion.
func asAnyMap(value any) (map[string]any, bool) {
	if m, ok := value.(map[string]any); ok {
		return m, true
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	for _, k := range rv.MapKeys() {
		out[k.String()] = rv.MapIndex(k).Interface()
	}
	return out, true
}

// asAnySlice mirrors asAnyMap for slices and arrays: fast path is a direct
// type assertion to []any, slow path reflects any slice/array element type
// into []any.
func asAnySlice(value any) ([]any, bool) {
	if s, ok := value.([]any); ok {
		return s, true
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

// dereference dereferences a pointer: only a single level of indirection is allowed.
func dereference(value any) (any, error) {
	for i := 0; ; i++ {
		if lo.IsNil(value) {
			return nil, nil
		}

		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Pointer && !val.IsNil() {
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
// ignoreUnknownKeys — when true, unknown fields are ignored (useful when ResourceData.Flatten sets values; new fields from the API can be safely skipped).
// setFlow — when true, objects must be wrapped in a list. See Schema.IsObject for more details.
func normalizeAny(sch *Schema, value any, ignoreUnknownKeys bool, setFlow bool) (any, error) {
	value, err := dereference(value)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return zeroValue(sch.Type), nil
	}

	// Preserved nil/unknown values from ModifyPlan are not real data; treat
	// them as missing so Expand / tfValue don't trip on the raw tftypes.Value.
	// Mirrors the same special case in getOk.
	if _, ok := value.(tftypes.Value); ok {
		return zeroValue(sch.Type), nil
	}

	if sch.Type.IsPrimitive() {
		str := fmt.Sprint(value)
		var v any
		var err error
		switch sch.Type {
		case SchemaTypeString:
			if str == "" && setFlow {
				// We don't want to store empty strings in the state
				return nil, nil
			}
			return str, nil
		case SchemaTypeInt:
			v, err = strconv.Atoi(str)
		case SchemaTypeFloat:
			v, err = strconv.ParseFloat(str, 64)
		case SchemaTypeBool:
			v, err = strconv.ParseBool(str)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse %q as %s: %w", str, sch.Type, err)
		}
		return v, nil
	}

	switch sch.Type {
	case SchemaTypeSet, SchemaTypeList:
		list, ok := asAnySlice(value)
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
			item, err := normalizeAny(sch.Items, elem, ignoreUnknownKeys, setFlow)
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
		dict, ok := asAnyMap(value)
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", value)
		}
		norm := make(map[string]any, len(dict))
		for key, elem := range dict {
			item, err := normalizeAny(sch.Items, elem, ignoreUnknownKeys, setFlow)
			if err != nil {
				return nil, err
			}
			norm[key] = item
		}
		return norm, nil
	case SchemaTypeObject:
		dict, ok := asAnyMap(value)
		if !ok {
			return nil, fmt.Errorf("expected object, got %T", value)
		}
		norm := make(map[string]any, len(dict))
		for key, elem := range dict {
			prop, ok := sch.Properties[key]
			if !ok {
				if ignoreUnknownKeys {
					// This is likely a new field from the API or a field we do not want to expose to the user.
					// Terraform might complain about unknown properties.
					continue
				}
				return nil, fmt.Errorf("unknown property %q", key)
			}
			item, err := normalizeAny(prop, elem, ignoreUnknownKeys, setFlow)
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

// isEmpty checks if the value is empty — has length zero.
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}

	return false
}
