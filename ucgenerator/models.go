package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
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

	IsDeprecated      bool   `yaml:"is_deprecated"`
	DeprecationNotice string `yaml:"deprecation_notice"`

	// Default value behaves as diff suppressor and messes things around.
	// For instance, if a bool field with value "true" and default value "false" is removed, GetOk() goes wild.
	// It can't tell if "false" is user's "false" or default "false".
	// So terraform just tries to set "false" instead of "true".
	// We don't use default values in schemas, but expose them in docs.
	// Because it makes things easier to control in diff suppressors.
	Default any `yaml:"default"`

	Enum []*struct {
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

var reCleanTFName = regexp.MustCompile(`[^a-z0-9_]`)

func (o *object) init(name string) {
	o.camelName = toCamelCase(name)
	o.varName = o.camelName + "Var"
	o.attrsName = o.camelName + "Attrs"
	o.tfoStructName = "tfo" + o.camelName
	o.dtoStructName = "dto" + o.camelName

	// If this one is a clone, it inherits the original jsonName
	if o.jsonName == "" {
		o.jsonName = name
	}

	if strings.HasPrefix(name, "pg_partman_bgw") || strings.HasPrefix(name, "pg_stat_") {
		// Legacy fields
		o.tfName = strings.ReplaceAll(name, ".", "__dot__")
	} else {
		o.tfName = reCleanTFName.ReplaceAllString(name, "_")
	}

	unwrapArrayMultipleTypes(o)

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
		child.init(k)
	}

	// Types can be list of strings, or a string
	if v, ok := o.OrigType.(string); ok {
		o.Type = objectType(v)
	} else if v, ok := o.OrigType.([]interface{}); ok {
		types := 0
		for _, t := range v {
			switch s := t.(string); s {
			case "null":
				o.Nullable = true
			default:
				o.Type = objectType(s)
				types++
				if types > 1 {
					log.Fatalf("%q has multiple types", name)
				}
			}
		}
	}

	if o.isArray() {
		o.ArrayItems.parent = o
		o.ArrayItems.init(name)
	}

	// In terraform objects are lists of one item.
	// So we need to add a constraint
	if o.isObject() {
		one := 1
		o.MaxItems = &one
		o.Default = nil
	}

	if o.isArray() && o.ArrayItems.isObject() {
		// In terraform object is a list of one object.
		// So a real list with one object is the same.
		// We need to see the difference for convert values to API
		if o.MaxItems != nil && *o.MaxItems == 1 {
			// As a fix, set nil to MaxItems
			log.Fatalf("%q array with object element and MaxItems==1", name)
		}
	}

	// A fix that removes empty string default value
	if o.Type == objectTypeString && o.Default != nil && o.Default.(string) == "" {
		o.Default = nil
	}

	// Sorts version enum values
	if o.Enum != nil && strings.HasSuffix(name, "version") {
		sort.Slice(o.Enum, func(i, j int) bool {
			return o.Enum[i].Value < o.Enum[j].Value
		})
	}
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

func (o *object) isSchemaless() bool {
	return o.isObject() && len(o.Properties) == 0
}

func (o *object) isArray() bool {
	return o.Type == objectTypeArray
}

func (o *object) isScalar() bool {
	return !(o.isObject() || o.isArray())
}

func (o *object) path() string {
	// Skips root name, i.e., pg_user_config
	// because it has dynamic prefix.
	if o.isRoot || o.parent.isRoot {
		return o.tfName
	}

	if o.parent.isArray() {
		// Array items should not appear in the path,
		// because technically they are indexes: foo.0.bar <- 0 is array item
		return o.parent.path()
	}
	return fmt.Sprintf("%s/%s", o.parent.path(), o.tfName)
}

// jsonPath same as path() but for jsonName
func (o *object) jsonPath() string {
	if o.isRoot || o.parent.isRoot {
		return o.jsonName
	}

	if o.parent.isArray() {
		return o.parent.jsonPath()
	}
	return fmt.Sprintf("%s/%s", o.parent.jsonPath(), o.jsonName)
}

// toCamelCase some fields have dots within, makes cleaner camelCase
func toCamelCase(s string) string {
	return strcase.LowerCamelCase(strings.ReplaceAll(s, ".", "_"))
}

func toUpperFirst(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func deepcopy(o *object) *object {
	clone := new(object)
	b, _ := json.Marshal(o)
	_ = json.Unmarshal(b, clone)
	return clone
}

const deprecationNotice = "Deprecated. Use `%s` instead."

// unwrapArrayMultipleTypes automatically unwraps multiple types ("type" as list or oneOf) for arrays.
// A "foo" field with types "string" and "object" unwrapped to three fields: foo, foo_string, foo_object
// First seen type becomes the default one and marked as deprecated.
// Because that's what happens to multi-typed fields:
// first they have one type, then new added, we split them into separate fields in terraform
// and deprecate the original field.
func unwrapArrayMultipleTypes(o *object) {
	for key, p := range o.Properties {
		// So far, array types unwrapped only
		if p.ArrayItems == nil {
			continue
		}

		prefix := key + "_"
		fields := make(map[string]*object)

		// Unwraps multiple _type names_, e.g. [string, object]
		types, ok := p.ArrayItems.OrigType.([]interface{})
		if ok {
			strTypes := make([]string, 0)
			for _, t := range types {
				if s := t.(string); s != "null" {
					strTypes = append(strTypes, s)
				}
			}

			if len(strTypes) == 1 {
				continue
			}

			// Multiple types.
			// This ArrayItems object is composite:
			// it has properties for the object type, and MaxLength for the string type.
			// So it just copies it and sets type explicitly.
			for _, s := range strTypes {
				clone := deepcopy(p)
				clone.jsonName = key
				clone.ArrayItems.OrigType = s
				fields[prefix+s] = clone
			}

			p.IsDeprecated = true
			p.DeprecationNotice = fmt.Sprintf(deprecationNotice, prefix+strTypes[0])
			p.ArrayItems.OrigType = strTypes[0]
			fields[key] = p

		} else if len(p.ArrayItems.OneOf) != 0 {
			// Unwraps multiple _type objects_, e.g. [{type:string}, {type: object}]
			for i := range p.ArrayItems.OneOf {
				t := p.ArrayItems.OneOf[i]
				clone := deepcopy(p)
				clone.jsonName = key
				clone.ArrayItems = t
				clone.Description = fmt.Sprintf("%s %s", addDot(p.Description), t.Description)
				fields[prefix+t.OrigType.(string)] = clone
			}

			// First seen type in priority. Replaces the original object
			priorityType := prefix + p.ArrayItems.OneOf[0].OrigType.(string)
			orig := deepcopy(fields[priorityType])
			orig.DeprecationNotice = fmt.Sprintf(deprecationNotice, priorityType)
			orig.IsDeprecated = true
			fields[key] = orig
		}

		for k, c := range fields {
			o.Properties[k] = c
		}
	}
}
