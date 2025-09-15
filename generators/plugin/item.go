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
	OperationUpsert Operation = "upsert"
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
	Schema      map[string]*Item     `yaml:"schema,omitempty"`
	Delete      []string             `yaml:"delete,omitempty"`
	Rename      map[string]string    `yaml:"rename,omitempty"`
	ObjectKey   string               `yaml:"objectKey,omitempty"`
	Resource    *SchemaMeta          `yaml:"resource,omitempty"`
	Datasource  *SchemaMeta          `yaml:"datasource,omitempty"`
	IDAttribute *IDAttribute         `yaml:"idAttribute"`
	Operations  map[string]Operation `yaml:"operations"`
}

type Item struct {
	Parent              *Item     `yaml:"-"`
	AppearsIn           AppearsIn `yaml:"appearsIn"`
	Name                string    `yaml:"name"`
	InIDAttribute       bool      `yaml:"inIDAttribute"`       // If field is part of ID attribute
	IDAttributePosition int       `yaml:"idAttributePosition"` // Position in the ID field

	// Tagged fields are exposed to the Definition.yaml
	Properties map[string]*Item `yaml:"properties"`
	Items      *Item            `yaml:"items"` // Array item or Map item

	// User-defined fields for YAML generation
	UserRequired  *bool `yaml:"required"`
	UserComputed  *bool `yaml:"computed"`
	UserOptional  *bool `yaml:"optional"`
	UserSensitive *bool `yaml:"sensitive"`
	UserForceNew  *bool `yaml:"forceNew"`

	// TF Validators
	// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/validators-predefined#background
	ConflictsWith []string `yaml:"conflictsWith"`
	ExactlyOneOf  []string `yaml:"exactlyOneOf"`
	AtLeastOneOf  []string `yaml:"atLeastOneOf"`
	AlsoRequires  []string `yaml:"alsoRequires"`

	JSONName           string     `yaml:"jsonName"`
	Type               SchemaType `yaml:"type"`
	Description        string     `yaml:"description"`
	DeprecationMessage string     `yaml:"deprecationMessage"`
	Required           bool       `yaml:"-"`
	Computed           bool       `yaml:"-"`
	Nullable           bool       `yaml:"-"`
	Optional           bool       `yaml:"-"`
	Sensitive          bool       `yaml:"-"`
	ForceNew           bool       `yaml:"-"`
	Default            any        `yaml:"default"`
	Enum               []any      `yaml:"enum"`
	MinLength          int        `yaml:"minLength"`
	MaxLength          int        `yaml:"maxLength"`
	MinItems           int        `yaml:"minItems"`
	MaxItems           int        `yaml:"maxItems"`
	Minimum            int        `yaml:"minimum"`
	Maximum            int        `yaml:"maximum"`
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

func (item *Item) ancestors() []*Item {
	seen := make(map[string]int)
	items := make([]*Item, 0)
	for v := item; v.Parent != nil; {
		// Duplicate check
		k := fmt.Sprint(v)
		seen[k]++
		if seen[k] > 1 {
			panic("Duplicate ancestor item found: " + v.Name)
		}

		// 1. Removes the root node, because it is the same for all items,
		//    and makes an unnecessary prefixing (v.Parent != nil)
		// 2. Ignores parent node when it is a map or array — they must have the same name,
		//    or the last part in the name will be duplicated
		if v.Parent.Items == nil {
			items = append(items, v)
		}
		v = v.Parent
	}

	slices.Reverse(items)
	return items
}

func (item *Item) JSONPath() string {
	chunks := make([]string, 0)
	for _, v := range item.ancestors() {
		chunks = append(chunks, v.JSONName)
	}
	return strings.Join(chunks, "/")
}

func (item *Item) Path() string {
	chunks := make([]string, 0)
	for _, v := range item.ancestors() {
		chunks = append(chunks, v.Name)
	}
	return strings.Join(chunks, "/")
}

func (item *Item) IsReadOnly(isResource bool) bool {
	if isResource {
		return !item.Required && !item.Optional
	}

	// ID attributes are not read-only in data sources
	return !item.InIDAttribute
}

func (item *Item) IsScalar() bool {
	switch item.Type {
	case SchemaTypeString, SchemaTypeInteger, SchemaTypeNumber, SchemaTypeBoolean:
		return true
	}
	return false
}

// IsNested is an object (not map) or an array with complex objects
func (item *Item) IsNested() bool {
	return item.IsObject() || item.IsArray() && item.Items.IsNested()
}

// IsMapNested returns true if the item is a map with complex objects
func (item *Item) IsMapNested() bool {
	return item.IsMap() && item.Items.IsNested()
}

func (item *Item) IsMap() bool {
	return item.Type == SchemaTypeObject && item.Items != nil
}

func (item *Item) IsArray() bool {
	return item.Type == SchemaTypeArray
}

func (item *Item) IsObject() bool {
	return item.Type == SchemaTypeObject && item.Items == nil && len(item.Properties) > 0
}

func (item *Item) IsRoot() bool {
	return item.Parent == nil
}

func (item *Item) IsRootProperty() bool {
	return !item.IsRoot() && item.Parent.IsRoot()
}

func (item *Item) IsEnum() bool {
	return len(item.Enum) > 0
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
