package main

import (
	"fmt"
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

// Content represents a request or response body.
type Content struct {
	ApplicationJSON struct {
		Schema struct {
			Ref string `json:"$ref"`
		} `json:"schema"`
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
	// SchemaTypeObject represents an object schema type.
	SchemaTypeObject SchemaType = "object"
	// SchemaTypeArray represents an array schema type.
	SchemaTypeArray SchemaType = "array"
	// SchemaTypeString represents a string schema type.
	SchemaTypeString SchemaType = "string"
	// SchemaTypeInteger represents an integer schema type.
	SchemaTypeInteger SchemaType = "integer"
	// SchemaTypeNumber represents a number schema type.
	SchemaTypeNumber SchemaType = "number"
	// SchemaTypeBoolean represents a boolean schema type.
	SchemaTypeBoolean SchemaType = "boolean"
)

// OASchema represents a parsed OpenAPIDoc schema.
type OASchema struct {
	Type                 SchemaType           `json:"type"`
	Properties           map[string]*OASchema `json:"properties"`
	AdditionalProperties *OASchema            `json:"additionalProperties"`
	Items                *OASchema            `json:"items"`
	Required             []string             `json:"required"`
	Enum                 []any                `json:"enum"`
	Default              any                  `json:"default"`
	MinItems             int                  `json:"minItems"`
	MaxItems             int                  `json:"maxItems"`
	MinLength            int                  `json:"minLength"`
	MaxLength            int                  `json:"maxLength"`
	Minimum              int                  `json:"minimum"`
	Maximum              int                  `json:"maximum"`
	Ref                  string               `json:"$ref"`
	Description          string               `json:"description"`
	CreateOnly           bool                 `json:"createOnly"`
	Sensitive            bool                 `json:"_secure"`
	Nullable             bool                 `json:"nullable"`
	name                 string               // json field name
}

// anyField The Aiven spec uses "any" as a property to render the "additionalProperties".
const anyField = "ANY"
