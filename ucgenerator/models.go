package main

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"
	"golang.org/x/exp/slices"
)

type objectType string

const (
	objectTypeObject  objectType = "object"
	objectTypeArray   objectType = "array"
	objectTypeString  objectType = "string"
	objectTypeBoolean objectType = "boolean"
	objectTypeInteger objectType = "integer"
	objectTypeNumber  objectType = "number"
)

type object struct {
	isRoot        bool   // top level object
	jsonName      string // original name from json spec
	tfName        string // terraform manifest field, unlike jsonName, can't store dot symbol
	tfoStructName string
	dtoStructName string
	camelName     string
	varName       string
	attrsName     string
	properties    []*object
	parent        *object

	Type     objectType `yaml:"-"`
	Required bool       `yaml:"-"`

	IsDeprecated bool `yaml:"is_deprecated"`
	Default      any  `yaml:"default"`
	Enum         []*struct {
		Value        string `yaml:"value"`
		IsDeprecated bool   `yaml:"is_deprecated"`
	} `yaml:"enum"`
	Pattern        string             `yaml:"pattern"`
	MinItems       *int               `yaml:"min_items"`
	MaxItems       *int               `yaml:"max_items"`
	MinLength      *int               `yaml:"min_length"`
	MaxLength      *int               `yaml:"max_length"`
	Minimum        *float64           `yaml:"minimum"`
	Maximum        *float64           `yaml:"maximum"`
	OrigType       any                `yaml:"type"`
	Format         string             `yaml:"format"`
	Title          string             `yaml:"title"`
	Description    string             `yaml:"description"`
	Properties     map[string]*object `yaml:"properties"`
	ArrayItems     *object            `yaml:"items"`
	OneOf          []*object          `yaml:"one_of"`
	RequiredFields []string           `yaml:"required"`
	CreateOnly     bool               `yaml:"create_only"`
	Nullable       bool               `yaml:"-"`
}

func (o *object) init(name string, seenObjs map[string]int) {
	// Sorts properties, so they keep order on each generation
	keys := make([]string, 0, len(o.Properties))
	for k := range o.Properties {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	for _, k := range keys {
		o.properties = append(o.properties, o.Properties[k])
	}

	required := make(map[string]bool, len(o.RequiredFields))
	for _, k := range o.RequiredFields {
		required[k] = true
	}

	for _, k := range keys {
		child := o.Properties[k]
		child.parent = o
		child.Required = required[k]
		child.init(k, seenObjs)
	}

	// Types can be list of strings, or a string
	if v, ok := o.OrigType.(string); ok {
		o.Type = objectType(v)
	} else if v, ok := o.OrigType.([]interface{}); ok {
		o.Type = objectType(v[0].(string))
		for _, t := range v {
			switch s := t.(string); s {
			case "null":
				o.Nullable = true
			default:
				o.Type = objectType(s)
			}
		}
	}

	// Arrays might have multiple types for its items.
	// But as a convention, they are never mixed
	// Terraform doesn't support multiple types,
	// so this takes objects over scalars as an assumption that the object is a newer type
	// — the extended version of the scalar
	if o.isArray() {
		for _, item := range o.ArrayItems.OneOf {
			o.ArrayItems = item

			// Object over another type
			if len(item.Properties) != 0 {
				break
			}
		}

		o.ArrayItems.parent = o
		o.ArrayItems.init(name, seenObjs)
	}

	// In terraform objects are lists of one item
	// Root item and properties should have max constraint
	if o.isObject() {
		if o.isRoot || o.parent != nil && o.parent.isObject() {
			o.MaxItems = toPtr(1)
		}
	}

	// Name depends on:
	// For objects it needs to resolve collision
	// For arrays it takes the item's name, which if is an object might have fixed name.
	o.camelName = toCamelCase(name)
	if o.isObject() {
		// Resolves name collision
		seenObjs[name]++
		if seenObjs[name] != 1 {
			o.camelName = fmt.Sprintf("%s%d", o.camelName, seenObjs[name])
		}
	}

	if o.isArray() && o.ArrayItems.isObject() {
		// Replaces with fixed name (no collision)
		o.camelName = o.ArrayItems.camelName
	}

	low := toLowerFirst(o.camelName)
	o.varName = low + "Var"
	o.attrsName = low + "Attrs"
	o.tfoStructName = "tfo" + o.camelName
	o.dtoStructName = "dto" + o.camelName
	o.jsonName = name
	o.tfName = strings.ReplaceAll(name, ".", "__")
}

func (o *object) isNestedBlock() bool {
	switch o.Type {
	case objectTypeObject:
		return len(o.Properties) > 0
	case objectTypeArray:
		return o.ArrayItems.isObject() || o.ArrayItems.isArray()
	}
	return false
}

func (o *object) isObject() bool {
	return o.Type == objectTypeObject
}

func (o *object) isArray() bool {
	return o.Type == objectTypeArray
}

func (o *object) isScalar() bool {
	return !(o.isObject() || o.isArray())
}

func (o *object) ListProperties() []*object {
	if o.isArray() {
		return o.ArrayItems.properties
	}
	return o.properties
}

// toCamelCase some fields have dots within, makes cleaner camelCase
func toCamelCase(s string) string {
	return strcase.UpperCamelCase(strings.ReplaceAll(s, ".", "_"))
}

func toLowerFirst(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}
