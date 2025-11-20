package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ettle/strcase"
)

// AppearsIn is a bitmask for the field appearance (request, response, etc.)
type AppearsIn uint

const (
	CreateHandler AppearsIn = 1 << iota
	ReadHandler
	UpdateHandler
	DeleteHandler
	PathParameter
	RequestBody
	ResponseBody
)

func (a AppearsIn) Contains(other AppearsIn) bool {
	return other > 0 && a&other == other
}

func listSources() []AppearsIn {
	return []AppearsIn{
		PathParameter,
		RequestBody,
		ResponseBody,
	}
}

type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationRead   OperationType = "read"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
	OperationUpsert OperationType = "upsert"
)

func listOperations() []OperationType {
	return []OperationType{
		OperationCreate,
		OperationRead,
		OperationUpdate,
		OperationDelete,
	}
}

func operationToHandler() map[OperationType]AppearsIn {
	return map[OperationType]AppearsIn{
		OperationCreate: CreateHandler,
		OperationRead:   ReadHandler,
		OperationUpdate: UpdateHandler,
		OperationDelete: DeleteHandler,
		OperationUpsert: CreateHandler | UpdateHandler,
	}
}

type OperationID string

type Operation struct {
	ID                   OperationID   `yaml:"id"`
	Type                 OperationType `yaml:"type"`
	DisableView          bool          `yaml:"disableView"`
	ResultKey            string        `yaml:"resultKey"`            // E.g.: {errors: [], result: {}} - extract "result"
	ResultListLookupKeys []string      `yaml:"resultListLookupKeys"` // When the response is a list, these keys are used to locate the correct item
	// When the API response is not an object (e.g., a primitive or array), wrap it into a map using this key.
	ResultToKey string `yaml:"resultToKey"`
}

type Operations []Operation

func (o Operations) ByID(operationID OperationID) Operation {
	for _, operation := range o {
		if operation.ID == operationID {
			return operation
		}
	}
	return Operation{}
}

func (o Operations) OperationIDTypes() map[OperationID]OperationType {
	result := make(map[OperationID]OperationType)
	for _, operation := range o {
		result[operation.ID] = operation.Type
	}
	return result
}

// AppearsInID returns the AppearsIn bitmask for a given operation ID and source (such as PathParameter, RequestBody, or ResponseBody).
// Example: o.AppearsInID("FooReadOperationID", PathParameter) will include:
//  1. The handler bit (e.g., ReadHandler) for this operation,
//  2. The source bit (e.g., PathParameter),
//  3. A unique bit specific to this operation ID and source, allowing fine-grained distinction.
//
// This enables querying for all sources of a type (like all path parameters),
// or isolating those associated with a particular operation ID.
// Different operations may define different sets of parameters, hence we need to distinguish them.
func (o Operations) AppearsInID(operationID OperationID, source AppearsIn) AppearsIn {
	sources := listSources()                // PathParameter, RequestBody, ResponseBody
	handlers := operationToHandler()        // CreateHandler, ReadHandler, etc.
	generic := len(handlers) + len(sources) // Reserved bits for generic bits: sources and handlers

	// Each operation has its own bucket of bits: [create:..read:...update:...delete:...]
	operationIDTypes := o.OperationIDTypes()
	operationBucket := len(sources) * slices.Index(sortedKeys(operationIDTypes), operationID)

	// Offset: [generic...create...read:path...update...delete...]
	// ---------------------------------^
	bitOffset := generic + operationBucket + slices.Index(sources, source)
	if bitOffset > 63 {
		panic(fmt.Sprintf("bitOffset overflow %s: generic=%d, bitOffset=%d", operationID, generic, bitOffset))
	}

	// E.g.: [generic(CreateHandler, ResponseBody), create(OperationSpecificResponse)]
	return handlers[operationIDTypes[operationID]] | source | (1 << bitOffset)
}

// AppearsInHandler finds all operation IDs matching the operation, merges the result
// There are can be multiple read operations, for example.
func (o Operations) AppearsInHandler(handler, source AppearsIn) AppearsIn {
	handlers := operationToHandler()
	var appearsIn AppearsIn
	for opID, opType := range o.OperationIDTypes() {
		if handlers[opType]&handler > 0 {
			appearsIn |= o.AppearsInID(opID, source)
		}
	}
	return appearsIn
}

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
	Fields      []string `yaml:"fields"`
	Description string   `yaml:"description,omitempty"`
	Mutable     bool     `yaml:"mutable,omitempty"` // Negative for UseStateForUnknown
}

type Definition struct {
	fileName        string            // e.g. organization_address.yaml
	typeName        string            // e.g. aiven_organization_address, aiven_kafka_topic
	Beta            bool              `yaml:"beta"`
	Location        string            `yaml:"location"`
	Schema          map[string]*Item  `yaml:"schema,omitempty"`
	Remove          []string          `yaml:"remove,omitempty"`
	Rename          map[string]string `yaml:"rename,omitempty"`
	Resource        *SchemaMeta       `yaml:"resource,omitempty"`
	Datasource      *SchemaMeta       `yaml:"datasource,omitempty"`
	IDAttribute     *IDAttribute      `yaml:"idAttribute"`
	LegacyTimeouts  bool              `yaml:"legacyTimeouts,omitempty"`
	Operations      Operations        `yaml:"operations"`
	Version         *int              `yaml:"version"`
	ClientHandler   string            `yaml:"clientHandler,omitempty"`
	RefreshState    bool              `yaml:"refreshState,omitempty"`
	ExpandModifier  bool              `yaml:"expandModifier,omitempty"`
	FlattenModifier bool              `yaml:"flattenModifier,omitempty"`
}

type Item struct {
	Parent              *Item     `yaml:"-"`
	AppearsIn           AppearsIn `yaml:"-"` // In Create, Read, Update, Delete request or response
	Name                string    `yaml:"-"`
	InIDAttribute       bool      `yaml:"-"` // If field is part of ID attribute
	IDAttributePosition int       `yaml:"-"` // Position in the ID field

	// Tagged fields are exposed to the Definition.yaml
	Properties map[string]*Item `yaml:"properties"`
	Items      *Item            `yaml:"items"` // Array item or Map item

	// Inherited from the user definition and propagated to "Items";
	// this simplifies generation logic, as both arrays and maps use "Items" similarly.
	AdditionalProperties *Item `yaml:"additionalProperties"`

	// User-defined fields for YAML generation
	OverrideRequired   *bool `yaml:"required"`
	OverrideComputed   *bool `yaml:"computed"`
	OverrideOptional   *bool `yaml:"optional"`
	OverrideSensitive  *bool `yaml:"sensitive"`
	OverrideForceNew   *bool `yaml:"forceNew"`
	UseStateForUnknown bool  `yaml:"useStateForUnknown"`

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
		// todo: replace with object in v5.0.0.
		return "List"
	case item.IsList():
		return "List"
	case item.IsSet():
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
	return item.IsSet() || item.IsList()
}

func (item *Item) IsSet() bool {
	return item.Type == SchemaTypeArray
}

func (item *Item) IsList() bool {
	return item.Type == SchemaTypeArrayOrdered
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
