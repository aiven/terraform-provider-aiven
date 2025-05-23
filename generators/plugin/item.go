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

type Definition struct {
	Beta        bool                 `yaml:"beta"`
	Location    string               `yaml:"location"`
	Schema      map[string]Item      `yaml:"schema,omitempty"`
	Delete      []string             `yaml:"delete,omitempty"`
	Rename      map[string]string    `yaml:"rename,omitempty"`
	ObjectKey   string               `yaml:"objectKey,omitempty"`
	Resource    *SchemaMeta          `yaml:"resource,omitempty"`
	Datasource  *SchemaMeta          `yaml:"datasource,omitempty"`
	IDAttribute *IDAttribute         `yaml:"idAttribute"`
	Operations  map[string]Operation `yaml:"operations"`
}

type Item struct {
	Parent              *Item
	AppearsIn           AppearsIn
	Name                string
	InIDAttribute       bool // If field is part of ID attribute
	IDAttributePosition int  // Position in the ID field

	// Tagged fields are exposed to the Definition.yaml
	JSONName           string           `yaml:"jsonName,omitempty"`
	Type               SchemaType       `yaml:"type,omitempty"`
	Description        string           `yaml:"description,omitempty"`
	DeprecationMessage string           `yaml:"deprecationMessage,omitempty"`
	Properties         map[string]*Item `yaml:"properties,omitempty"`
	Items              *Item            `yaml:"items,omitempty"` // Array item or Map item
	Required           bool             `yaml:"required,omitempty"`
	Computed           bool             `yaml:"computed,omitempty"`
	Nullable           bool             `yaml:"nullable,omitempty"`
	Optional           bool             `yaml:"optional,omitempty"`
	Sensitive          bool             `yaml:"sensitive,omitempty"`
	Default            any              `yaml:"default,omitempty"`
	Enum               []any            `yaml:"enum,omitempty"`
	ForceNew           bool             `yaml:"forceNew,omitempty"`
	MinLength          int              `yaml:"minLength,omitempty"`
	MaxLength          int              `yaml:"maxLength,omitempty"`
	MinItems           int              `yaml:"minItems,omitempty"`
	MaxItems           int              `yaml:"maxItems,omitempty"`
	Minimum            int              `yaml:"minimum,omitempty"`
	Maximum            int              `yaml:"maximum,omitempty"`
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
		v := item.Items.ApiType()
		if item.Items.IsScalar() {
			// Remove the pointer for scalar types
			v = v[1:]
		}
		return "[]" + v
	}

	if item.IsMap() {
		return "map[string]" + item.Items.ApiType()
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
	return item.IsObject() || item.IsArray() && item.Items.IsObject()
}

func (item *Item) IsMap() bool {
	return item.Type == SchemaTypeObject && item.Items != nil
}

func (item *Item) IsArray() bool {
	return item.Type == SchemaTypeArray
}

func (item *Item) IsObject() bool {
	return item.Type == SchemaTypeObject && item.Items == nil
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
func mergeItems(a, b *Item, override bool) error {
	if a == nil {
		return nil
	}

	if b == nil {
		return nil
	}

	// JSON names are not allowed to be different, until it's not in the definition.yaml
	if a.JSONName != b.JSONName && len(a.JSONName)*len(b.JSONName) > 0 {
		return fmt.Errorf("field %q JSON names do not match: %s != %s", a.Path(), a.JSONName, b.JSONName)
	}

	switch {
	case override:
		// The definition.yaml has priority over the API spec
		a.Type = or(b.Type, a.Type)

		// Invalidates the fields that are not compatible with the new type
		switch a.Type {
		case SchemaTypeObject:
			a.Items = nil
		case SchemaTypeArray:
			a.Properties = nil
		default:
			a.Items = nil
			a.Properties = nil
		}

		// Sets the required and optional fields
		// Can't be both required and optional
		if b.Required || b.Optional {
			a.Required = b.Required || !b.Optional
			a.Optional = b.Optional || !b.Required
		}
	case a.Type != b.Type:
		return fmt.Errorf("field %q types do not match: %s != %s ", a.Path(), a.Type, b.Type)
	default:
		// Enums can be different for request and response
		a.Enum = mergeSlices(a.Enum, b.Enum)
		a.Required = a.Required || b.Required
		a.Optional = a.Optional || b.Optional
	}

	// Copies properties
	// TODO: there is no way to set Sensitive false, or Computed false, and so on.
	a.Description = or(a.Description, b.Description)
	a.DeprecationMessage = or(a.DeprecationMessage, b.DeprecationMessage)
	a.AppearsIn |= b.AppearsIn
	a.Nullable = a.Nullable || b.Nullable
	a.ForceNew = a.ForceNew || b.ForceNew
	a.Sensitive = a.Sensitive || b.Sensitive
	a.Computed = a.Computed || b.Computed

	// Validate that a field cannot be both computed and required
	if a.Computed && a.Required {
		return fmt.Errorf("field %q cannot be both computed and required", a.Path())
	}

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
	err = mergeItemProperties(a.Properties, b.Properties, override)
	if err != nil {
		return err
	}

	err = mergeItems(a.Items, b.Items, override)
	if err != nil {
		return err
	}

	// Sometimes OperationIDs have different descriptions.
	// In that case, there is a good chance that the longest one has the most details.
	if len(b.Description) > len(a.Description) {
		a.Description = b.Description
	}
	return nil
}

func mergeItemProperties(a, b map[string]*Item, override bool) error {
	for k, v := range b {
		if vv, ok := a[k]; ok {
			err := mergeItems(vv, v, override)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
