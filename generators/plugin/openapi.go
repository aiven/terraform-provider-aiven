package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
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

// resolveSchema resolves the ref and turns a "oneOf" into a free-form object,
// keeping its description and flags, e.g. "_secure".
func (d *OpenAPIDoc) resolveSchema(s *OASchema) (*OASchema, error) {
	s, err := d.resolveRef(s)
	if err != nil || s == nil || len(s.OneOf) == 0 {
		return s, err
	}
	// A copy: components are shared
	o := *s
	o.Type = SchemaTypeObject
	o.Properties, o.AdditionalProperties, o.Items, o.OneOf = nil, nil, nil, nil
	return &o, nil
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
	Description          string                  `json:"description"`
	CreateOnly           bool                    `json:"createOnly"`
	Sensitive            bool                    `json:"_secure"`
	Nullable             bool                    `json:"nullable"`
	name                 string                  // json field name
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
