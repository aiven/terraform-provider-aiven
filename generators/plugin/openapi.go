package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/samber/lo"
)

// OpenAPIDoc represents a parsed OpenAPI document.
type OpenAPIDoc struct {
	Paths      map[string]map[string]*OAPath `json:"paths"`
	Components struct {
		Schemas    map[string]*OASchema    `json:"schemas"`
		Parameters map[string]*OAParameter `json:"parameters"`
	} `json:"components"`
}

const componentIndex = 3

func (d *OpenAPIDoc) getParameterRef(path string) (*OAParameter, error) {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]

	param, ok := d.Components.Parameters[name]
	if !ok {
		return nil, fmt.Errorf("param %q not found", path)
	}
	return param, nil
}

func (d *OpenAPIDoc) getComponentRef(path string) (*OASchema, error) {
	chunks := strings.Split(path, "/")
	if len(chunks) < componentIndex+1 {
		return nil, fmt.Errorf("invalid schema path %q", path)
	}

	name := chunks[componentIndex]
	schema := d.Components.Schemas[name]
	for i := componentIndex + 1; i < len(chunks); i++ {
		switch k := chunks[i]; k {
		case "items":
			schema = schema.Items
		case "properties":
			schema = schema.Properties[chunks[i+1]]
			i++
		default:
			return nil, fmt.Errorf("unknown schema path %v: %s", chunks, k)
		}
	}

	if schema == nil {
		return nil, fmt.Errorf("schema %q not found", path)
	}

	schema.name = name
	return schema, nil
}

// resolveRef follows the ref chain to the component, or returns the schema itself.
func (d *OpenAPIDoc) resolveRef(s *OASchema) (*OASchema, error) {
	seen := make(map[string]bool)
	for s != nil && s.Ref != "" {
		if seen[s.Ref] {
			return nil, fmt.Errorf("cyclic schema ref %q", s.Ref)
		}
		seen[s.Ref] = true

		var err error
		s, err = d.getComponentRef(s.Ref)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// resolveSchema resolves the ref, merges the object variants, see mergeVariants,
// and turns any other "oneOf" into a free-form object, keeping its description and flags, e.g. "_secure".
func (d *OpenAPIDoc) resolveSchema(s *OASchema) (*OASchema, error) {
	s, err := d.resolveRef(s)
	if err == nil && s != nil {
		s, err = d.mergeVariants(s)
	}
	if err != nil || s == nil || len(s.OneOf) == 0 {
		return s, err
	}
	// A copy: components are shared
	o := *s
	o.Type = SchemaTypeObject
	o.Properties, o.AdditionalProperties, o.Items, o.OneOf = nil, nil, nil, nil
	return &o, nil
}

// schemaVariant is an object schema of a "oneOf" or "anyOf", with the JSON names.
type schemaVariant struct {
	title      string
	properties []string
	required   []string // Besides the properties the merged schema requires
}

// mergeVariants merges the object variants of a "oneOf", "anyOf" or "allOf" into one object.
// An "allOf" requires what any variant requires. A "oneOf" or "anyOf" requires what all variants require,
// and keeps the variants for the validator and the docs, see Item.Variants.
// Returns the schema itself when a variant isn't an object.
func (d *OpenAPIDoc) mergeVariants(s *OASchema) (*OASchema, error) {
	variants, allOf := s.OneOf, false
	switch {
	case len(s.OneOf) > 0:
	case len(s.AnyOf) > 0:
		variants = s.AnyOf
	case len(s.AllOf) > 0:
		variants, allOf = s.AllOf, true
	default:
		return s, nil
	}

	resolved := make([]*OASchema, len(variants))
	for i, v := range variants {
		v, err := d.resolveSchema(v)
		if err != nil {
			return nil, err
		}
		if v == nil || v.Type != SchemaTypeObject || len(v.Properties) == 0 {
			return s, nil
		}
		resolved[i] = v
	}

	// A copy: components are shared
	o := *s
	o.Type, o.OneOf, o.AnyOf, o.AllOf = SchemaTypeObject, nil, nil, nil
	o.Properties, o.AdditionalProperties = make(map[string]*OASchema), nil
	closed := true
	for _, v := range resolved {
		for k, p := range v.Properties {
			var err error
			o.Properties[k], err = d.mergeProperty(o.Properties[k], p)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
		}
		closed = closed && v.AdditionalProperties != nil && !v.AdditionalProperties.Allowed
	}
	if closed {
		o.AdditionalProperties = new(OAAdditionalProperties)
	}

	required := lo.Map(resolved, func(v *OASchema, _ int) []string { return v.Required })
	if allOf {
		o.Required = lo.Union(append(required, s.Required)...)
		return &o, nil
	}

	o.Required = lo.Union(s.Required, lo.Intersect(required...))
	o.anyVariant = len(s.OneOf) == 0
	for i, v := range resolved {
		o.variants = append(o.variants, schemaVariant{
			title:      cmp.Or(v.Title, fmt.Sprintf("Variant %d", i+1)),
			properties: sortedKeys(v.Properties),
			required:   lo.Without(v.Required, o.Required...),
		})
	}
	return &o, nil
}

// mergeProperty merges a property that several variants have. The variants must agree on it:
// one of them can miss the description, the title or the example, the rest must be equal.
// Different descriptions or titles fail later, unless the definition describes the field, see Item.DescriptionConflict.
func (d *OpenAPIDoc) mergeProperty(a, b *OASchema) (*OASchema, error) {
	b, err := d.resolveSchema(b)
	if err != nil || a == nil {
		return b, err
	}
	a, err = d.resolveSchema(a)
	if err != nil {
		return nil, err
	}

	o := *a
	set := func(v any) bool { return v != nil && v != "" }
	for _, f := range []struct {
		name string
		a, b any
	}{
		{"description", a.Description, b.Description},
		{"title", a.Title, b.Title},
		{"example", a.Example, b.Example},
	} {
		if !set(f.a) || !set(f.b) || reflect.DeepEqual(f.a, f.b) {
			continue
		}
		err := variantsDiffer(f.name, f.a, f.b)
		if f.name == "example" {
			return nil, err
		}
		o.descriptionConflict = cmp.Or(o.descriptionConflict, err.Error())
	}
	o.Description, o.Title = cmp.Or(a.Description, b.Description), cmp.Or(a.Title, b.Title)
	if o.Example == nil {
		o.Example = b.Example
	}

	// The fields besides the ones above and the nested schemas, which are merged below
	x, y := *a, *b
	for _, s := range []*OASchema{&x, &y} {
		s.Description, s.Title, s.Example, s.Ref, s.Properties, s.Items = "", "", nil, "", nil, nil
	}
	if lo.ElementsMatch(x.Required, y.Required) {
		y.Required = x.Required
	}
	if lo.ElementsMatch(x.Enum, y.Enum) {
		y.Enum = x.Enum
	}
	vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)
	for i := range vx.NumField() {
		f := vx.Type().Field(i)
		if f.IsExported() && !reflect.DeepEqual(vx.Field(i).Interface(), vy.Field(i).Interface()) {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			return nil, variantsDiffer(name, vx.Field(i).Interface(), vy.Field(i).Interface())
		}
	}
	if !reflect.DeepEqual(a.variants, b.variants) || a.anyVariant != b.anyVariant {
		return nil, errors.New("the variants have different oneOf or anyOf variants")
	}

	if (a.Items == nil) != (b.Items == nil) {
		return nil, variantsDiffer("items", a.Items != nil, b.Items != nil)
	}
	if a.Items != nil {
		o.Items, err = d.mergeProperty(a.Items, b.Items)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}
	}

	if !lo.ElementsMatch(sortedKeys(a.Properties), sortedKeys(b.Properties)) {
		return nil, variantsDiffer("properties", sortedKeys(a.Properties), sortedKeys(b.Properties))
	}
	if len(a.Properties) > 0 {
		o.Properties = make(map[string]*OASchema, len(a.Properties))
		for k, p := range a.Properties {
			o.Properties[k], err = d.mergeProperty(p, b.Properties[k])
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
		}
	}
	return &o, nil
}

func variantsDiffer(field string, a, b any) error {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return fmt.Errorf("the variants have different %s: %s and %s", field, x, y)
}

// Content represents a request or response body.
type Content struct {
	ApplicationJSON struct {
		Schema struct {
			Ref string `json:"$ref"`
		} `json:"schema"`
		Example map[string]any `json:"example"` // Contains a json object with example values for the schema
	} `json:"application/json"`
}

// OAPath represents a parsed OpenAPIDoc path.
type OAPath struct {
	Tags        []string       `json:"tags"`
	OperationID string         `json:"operationId"`
	Parameters  []*OAParameter `json:"parameters"`
	Summary     string         `json:"summary"`
	Deprecated  bool           `json:"deprecated"`
	Request     struct {
		Content Content `json:"content"`
	} `json:"requestBody"`
	Response struct {
		OK struct {
			Content Content `json:"content"`
		} `json:"200"`
		NoContent struct {
			Content Content `json:"content"`
		} `json:"204"`
	} `json:"responses"`
	Experimental bool `json:"x-experimental"`
}

// OAParameter represents a parsed OpenAPIDoc parameter.
type OAParameter struct {
	Ref         string    `json:"$ref"`
	In          string    `json:"in"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schema      *OASchema `json:"schema"`
	Required    bool      `json:"required"`
}

// SchemaType represents a schema type.
type SchemaType string

const (
	SchemaTypeObject       SchemaType = "object"
	SchemaTypeArray        SchemaType = "array"        // Rendered as TypeSet
	SchemaTypeArrayOrdered SchemaType = "arrayOrdered" // Rendered as TypeList
	SchemaTypeString       SchemaType = "string"
	SchemaTypeInteger      SchemaType = "integer"
	SchemaTypeNumber       SchemaType = "number"
	SchemaTypeBoolean      SchemaType = "boolean"
)

// StringFormat is a string schema format. Known ones get a validator, see stringFormats.
type StringFormat string

const (
	StringFormatCIDR     StringFormat = "cidr"
	StringFormatIPv4CIDR StringFormat = "ipv4-cidr"
	StringFormatUUID     StringFormat = "uuid"
)

// OASchema represents a parsed OpenAPIDoc schema.
type OASchema struct {
	Type                 SchemaType              `json:"type"`
	Properties           map[string]*OASchema    `json:"properties"`
	AdditionalProperties *OAAdditionalProperties `json:"additionalProperties"`
	Items                *OASchema               `json:"items"`
	Required             []string                `json:"required"`
	Enum                 []any                   `json:"enum"`
	Default              any                     `json:"default"`
	Example              any                     `json:"example"`
	MinItems             int                     `json:"minItems"`
	MaxItems             int                     `json:"maxItems"`
	MinLength            int                     `json:"minLength"`
	MaxLength            int                     `json:"maxLength"`
	Minimum              json.Number             `json:"minimum"`
	Maximum              json.Number             `json:"maximum"`
	Format               StringFormat            `json:"format"`
	Pattern              string                  `json:"pattern"`
	Ref                  string                  `json:"$ref"`
	OneOf                []*OASchema             `json:"oneOf"`
	AnyOf                []*OASchema             `json:"anyOf"`
	AllOf                []*OASchema             `json:"allOf"`
	Title                string                  `json:"title"`
	Description          string                  `json:"description"`
	CreateOnly           bool                    `json:"createOnly"`
	Sensitive            bool                    `json:"_secure"`
	Nullable             bool                    `json:"nullable"`

	// User config schemas prefix the extensions with "x-"
	CreateOnlyExt bool `json:"x-createOnly"`
	SensitiveExt  bool `json:"x-_secure"`

	name string // json field name

	// Set by mergeVariants
	variants            []schemaVariant
	anyVariant          bool
	descriptionConflict string

	// Set by splitMultiTypeArrays
	jsonName      string
	deprecation   string
	conflictsWith []string
}

// OAAdditionalProperties is either the schema of map values,
// or a boolean that allows or forbids any extra keys.
type OAAdditionalProperties struct {
	Schema  *OASchema
	Allowed bool
}

func (a *OAAdditionalProperties) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &a.Allowed); err == nil {
		return nil
	}
	a.Allowed = true
	return json.Unmarshal(b, &a.Schema)
}

// numberToInt converts a JSON number, rounding a fraction up or down.
// Clamps to the int range: out-of-range float to int conversion is platform-dependent.
func numberToInt(n json.Number, roundUp bool) int {
	if i, err := n.Int64(); err == nil {
		return int(i)
	}
	f, _ := n.Float64()
	if roundUp {
		f = math.Ceil(f)
	} else {
		f = math.Floor(f)
	}
	switch {
	case f >= math.MaxInt:
		return math.MaxInt
	case f <= math.MinInt:
		return math.MinInt
	}
	return int(f)
}

// anyField The Aiven spec uses "any" as a property to render the "additionalProperties".
const anyField = "ANY"
