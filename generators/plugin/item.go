package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ettle/strcase"
)

// AppearsIn is a bitmask for the field appearance (request, response, etc.)
type AppearsIn int

const (
	PathParameter AppearsIn = 1 << iota
	RequestBody
	ResponseBody
	CreateHandler
	ReadHandler
	UpdateHandler
	DeleteHandler
	CreatePathParameter
	CreateRequestBody
	UpdateRequestBody
	CreateResponseBody
)

type Operation string

const (
	OperationCreate Operation = "create"
	OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
)

type Scope struct {
	OpenAPI     *OpenAPIDoc
	Definition  *Definition
	CurrentMeta *SchemaMeta
}

// SchemaMeta contains fields to override.
// Extend this struct if you need to override more fields and update the usage.
type SchemaMeta struct {
	Description        string `yaml:"description"`
	DeprecationMessage string `yaml:"deprecationMessage,omitempty"`
}

type IDAttribute struct {
	Compose     []string `yaml:"compose"`
	Description string   `yaml:"description,omitempty"`
	Mutable     bool     `yaml:"mutable,omitempty"` // Negative for UseStateForUnknown
}

type Patch struct {
	Rename             string `yaml:"rename,omitempty"`
	Description        string `yaml:"description,omitempty"`
	DeprecationMessage string `yaml:"deprecationMessage,omitempty"`
	Sensitive          bool   `yaml:"sensitive,omitempty"`
	Optional           bool   `yaml:"optional,omitempty"`
	Required           bool   `yaml:"required,omitempty"`
	Computed           bool   `yaml:"computed,omitempty"`
	Delete             bool   `yaml:"delete,omitempty"`
}

type Definition struct {
	Beta        bool                 `yaml:"beta"`
	Location    string               `yaml:"location"`
	Patch       map[string]Patch     `yaml:"patch,omitempty"`
	ObjectKey   string               `yaml:"objectKey,omitempty"`
	Resource    *SchemaMeta          `yaml:"resource,omitempty"`
	Datasource  *SchemaMeta          `yaml:"datasource,omitempty"`
	IDAttribute *IDAttribute         `yaml:"idAttribute"`
	Operations  map[string]Operation `yaml:"operations"`
}

type Item struct {
	Parent               *Item
	Name                 string
	JSONName             string
	Type                 SchemaType
	Description          string
	DeprecationMessage   string
	Required             bool
	Properties           map[string]*Item
	Element              *Item // Array item or Map item, "element" is a TF naming convention
	InIDAttribute        bool  // If field is part of ID attribute
	IDAttributePosition  int   // Position in the ID field
	ForceNew             bool
	Computed             bool
	Optional             bool
	Sensitive            bool
	Nullable             bool
	Enum                 []any
	Default              any
	MinItems             int
	MaxItems             int
	MinLength            int
	MaxLength            int
	Minimum              int
	Maximum              int
	AppearsIn            AppearsIn
	AdditionalProperties *Item
}

// UniqueName generates unique name by composing all ancestor names
func (item *Item) UniqueName() string {
	p := strings.ReplaceAll(item.Path(), "/", "_")
	return firstUpper(strcase.ToGoCamel(p))
}

func (item *Item) GoVarName() string {
	return "v" + item.GoFieldName()
}

func (item *Item) GoFieldName() string {
	if item.Name == "id" {
		// Manually fix the ID field name
		return "ID"
	}
	return firstUpper(strcase.ToGoCamel(item.Name))
}

func (item *Item) TFModelName() string {
	if item.IsRoot() {
		return tfRootModel
	}
	return tfModelPrefix + item.UniqueName()
}

func (item *Item) ApiModelName() string {
	if item.IsRoot() {
		return apiRootModel
	}
	return apiModelPrefix + item.UniqueName()
}

func (item *Item) ApiType() string {
	if item.IsObject() {
		return "*" + item.ApiModelName()
	}

	if item.IsArray() {
		v := item.Element.ApiType()
		if item.Element.IsScalar() {
			// Remove the pointer for scalar types
			v = v[1:]
		}
		return "[]" + v
	}

	if item.IsMap() {
		return "map[string]" + item.Element.ApiType()
	}

	return "*" + strings.ToLower(item.TFType())
}

func (item *Item) TFType() string {
	switch {
	case item.IsObject():
		// List type is compatible with SDKv2
		return "List"
	case item.IsArray():
		return "Set"
	case item.IsMap():
		return "Map"
	}
	return typingMapping()[item.Type]
}

func (item *Item) JSONPath() string {
	parent := item.Parent
	chunks := []string{item.JSONName}
	for parent != nil {
		if !parent.IsArray() {
			chunks = append(chunks, parent.JSONName)
		}
		parent = parent.Parent
	}

	slices.Reverse(chunks)

	// Removes the root node, because it is the same for all items,
	// and makes an unnecessary prefixing
	return strings.Join(chunks[1:], "/")
}

func (item *Item) Path() string {
	parent := item.Parent
	chunks := []string{item.Name}
	for parent != nil {
		if !parent.IsArray() {
			chunks = append(chunks, parent.Name)
		}
		parent = parent.Parent
	}

	slices.Reverse(chunks)

	// Removes the root node, because it is the same for all items,
	// and makes an unnecessary prefixing
	return strings.Join(chunks[1:], "/")
}

func (item *Item) IsReadOnly() bool {
	// XOR, they both can be false, but only one can be true
	return item.Required == item.Optional
}

func (item *Item) IsScalar() bool {
	switch item.Type {
	case SchemaTypeArray, SchemaTypeObject:
		return false
	}
	return true
}

// IsNested is an object (not map) or an array with complex objects
func (item *Item) IsNested() bool {
	return item.IsObject() || item.IsArray() && item.Element.IsObject()
}

func (item *Item) IsMap() bool {
	return item.Type == SchemaTypeObject && item.Element != nil
}

func (item *Item) IsArray() bool {
	return item.Type == SchemaTypeArray
}

func (item *Item) IsObject() bool {
	return item.Type == SchemaTypeObject && item.Element == nil
}

func (item *Item) IsRoot() bool {
	return item.Parent == nil
}

func (item *Item) IsRootProperty() bool {
	return !item.IsRoot() && item.Parent.IsRoot()
}

func (item *Item) GetIDFields() []*Item {
	fields := make([]*Item, 0)
	for _, v := range item.Properties {
		if v.InIDAttribute {
			fields = append(fields, v)
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].IDAttributePosition < fields[j].IDAttributePosition
	})
	return fields
}

// mergeItems merges b into a
func mergeItems(a, b *Item) error {
	if a == nil {
		return nil
	}

	if b == nil {
		return nil
	}

	if a.Type != b.Type {
		return fmt.Errorf("field %q types do not match: %s != %s ", a.Path(), a.Type, b.Type)
	}

	// Enums can be different for request and response
	a.Enum = mergeSlices(a.Enum, b.Enum)

	// Copies properties
	a.AppearsIn |= b.AppearsIn
	a.Nullable = a.Nullable || b.Nullable
	a.Required = a.Required || b.Required
	a.ForceNew = a.ForceNew || b.ForceNew
	a.Sensitive = a.Sensitive || b.Sensitive
	a.Computed = a.Computed || b.Computed
	a.MinItems = or(a.MinItems, b.MinItems)
	a.MaxItems = or(a.MaxItems, b.MaxItems)
	a.MinLength = or(a.MinLength, b.MinLength)
	a.MaxLength = or(a.MaxLength, b.MaxLength)
	a.Minimum = or(a.Minimum, b.Minimum)
	a.Maximum = or(a.Maximum, b.Maximum)
	a.JSONName = or(a.JSONName, b.JSONName)

	if a.Default == nil {
		a.Default = b.Default
	}

	var err error
	a.Properties, err = mergeItemProperties(a.Properties, b.Properties)
	if err != nil {
		return err
	}

	err = mergeItems(a.Element, b.Element)
	if err != nil {
		return err
	}

	if len(b.Description) > len(a.Description) {
		a.Description = b.Description
	}
	return nil
}

func mergeItemProperties(a, b map[string]*Item) (map[string]*Item, error) {
	result := make(map[string]*Item)
	for k, v := range a {
		result[k] = v
	}

	for k, v := range b {
		if _, ok := result[k]; ok {
			err := mergeItems(result[k], v)
			if err != nil {
				return nil, err
			}
		} else {
			result[k] = v
		}
	}
	return result, nil
}
