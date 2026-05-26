package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

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
)

func listOperationTypes() []OperationType {
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
	}
}

type OperationID string

type Operation struct {
	ID                   OperationID       `yaml:"id"`
	Type                 OperationType     `yaml:"type"`
	DisableView          bool              `yaml:"disableView"`
	WaitForDeletion      bool              `yaml:"waitForDeletion"`
	DatasourceLookup     bool              `yaml:"datasourceLookup"`     // Inline this read op as the data source readView's id-empty branch (lookup by alternative key). Not wired into resource views.
	ResultIDField        string            `yaml:"resultIDField"`        // Go field name on the lookup result item that holds the primary id. When set on a datasourceLookup op, the lookup resolves the id and control falls through to the canonical read body in the same readView.
	ResultKey            string            `yaml:"resultKey"`            // E.g.: {errors: [], result: {}} - extract "result"
	ResultKeyField       string            `yaml:"resultKeyField"`       // Go field name on the client response struct to drill into before further processing. Combined with resultListLookupKeys it points at the list to search; otherwise it points at the inner object to flatten. E.g.: "Aws" emits d.Flatten(rsp.Aws).
	ResultListLookupKeys map[string]string `yaml:"resultListLookupKeys"` // When the response is a list, these keys are used to locate the correct item
	// When the API response is not an object (e.g., a primitive or array), wrap it into a map using this key.
	ResultToKey string `yaml:"resultToKey"`

	Request  *OASchema
	Response *OASchema
}

type Operations []*Operation

// AppearsInID returns the AppearsIn bitmask for a given operation ID and handler type and source (such as PathParameter, RequestBody, or ResponseBody).
// Example: o.AppearsInID("FooReadOperationID", ReadHandler, PathParameter) will include:
//  1. The handler bit (e.g., ReadHandler) for this operation,
//  2. The source bit (e.g., PathParameter),
//  3. A unique bit specific to this operation ID and handler type and source, allowing fine-grained distinction.
//
// This enables querying for all sources of a type (like all path parameters) for a given operation ID and handler type,
// or isolating those associated with a particular operation ID and handler type.
// Different operations may define different sets of parameters, hence we need to distinguish them.
func (o Operations) AppearsInID(operationID OperationID, handler OperationType, source AppearsIn) AppearsIn {
	operationIndex := slices.IndexFunc(o, func(op *Operation) bool {
		return op.ID == operationID && op.Type == handler
	})

	sources := listSources()                            // PathParameter, RequestBody, ResponseBody
	generic := len(sources) + len(listOperationTypes()) // Reserved bits for generic bits: sources and generic handlers

	// Offset: [generic...create...read:path...update...delete...]
	// ---------------------------------^
	operationBucket := operationIndex * len(sources)
	bitOffset := generic + operationBucket + slices.Index(sources, source)
	if bitOffset > 63 {
		panic(fmt.Sprintf("bitOffset overflow %s: generic=%d, bitOffset=%d", operationID, generic, bitOffset))
	}

	// E.g.: [generic(CreateHandler, ResponseBody), create(OperationSpecificResponse)]
	return operationToHandler()[o[operationIndex].Type] | source | (1 << bitOffset)
}

// AppearsInHandler finds all operation IDs matching the operation, merges the result
// There are can be multiple read operations, for example.
func (o Operations) AppearsInHandler(handler OperationType, source AppearsIn) AppearsIn {
	var appearsIn AppearsIn
	for _, op := range o {
		if op.Type == handler {
			appearsIn |= o.AppearsInID(op.ID, handler, source)
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
	Description           string        `yaml:"description"`
	DeprecationMessage    string        `yaml:"deprecationMessage,omitempty"`
	TerminationProtection bool          `yaml:"terminationProtection,omitempty"`
	RefreshState          bool          `yaml:"refreshState,omitempty"`
	RefreshStateDelay     time.Duration `yaml:"refreshStateDelay,omitempty"`
	RefreshStateWaiter    bool          `yaml:"refreshStateWaiter,omitempty"`
	RemoveMissing         bool          `yaml:"removeMissing,omitempty"`
	IgnoreAlreadyExists   bool          `yaml:"ignoreAlreadyExists,omitempty"`
	DisableExample        bool          `yaml:"disableExample,omitempty"`
	ValidateConfig        bool          `yaml:"validateConfig,omitempty"`
	ModifyPlan            bool          `yaml:"modifyPlan,omitempty"`

	// SchemaOverride is a datasource-only schema overlay merged on top of the
	// base Definition.Schema when generating the datasource. Resource
	// generation ignores this field. Only meaningful on Definition.Datasource.
	SchemaOverride map[string]*Item `yaml:"schemaOverride,omitempty"`

	// ExactlyOneOf declares the datasource's top-level "exactly one of"
	// discriminator group. Drives the example LOOKUP hint and the
	// `datasourcevalidator.ExactlyOneOf` block in datasourceConfigValidators.
	ExactlyOneOf []string `yaml:"exactlyOneOf,omitempty"`
}

type Definition struct {
	fileName            string            // e.g. organization_address.yaml
	typeName            string            // e.g. aiven_organization_address, aiven_kafka_topic
	Beta                *bool             `yaml:"beta"` // Is figured as beta from `x-experimental` OpenAPI field
	LimitedAvailability *bool             `yaml:"limitedAvailability"`
	Location            string            `yaml:"location"`
	Schema              map[string]*Item  `yaml:"schema,omitempty"`
	Remove              []string          `yaml:"remove,omitempty"`
	Rename              map[string]string `yaml:"rename,omitempty"`
	Resource            *SchemaMeta       `yaml:"resource,omitempty"`
	Datasource          *SchemaMeta       `yaml:"datasource,omitempty"`
	IDAttributeComposed []string          `yaml:"idAttributeComposed,omitempty"`
	LegacyTimeouts      bool              `yaml:"legacyTimeouts,omitempty"`
	Operations          Operations        `yaml:"operations"`
	Version             *int              `yaml:"version"`
	ClientHandler       string            `yaml:"clientHandler,omitempty"`
	PlanModifier        bool              `yaml:"planModifier,omitempty"`
	ExpandModifier      bool              `yaml:"expandModifier,omitempty"`
	FlattenModifier     bool              `yaml:"flattenModifier,omitempty"`
}

// DatasourceLookupOp returns the single read op marked `datasourceLookup`, or nil. It
// is the authoritative source for the data source's lookup contract (validators,
// schema flags, example hint, readView's id-empty branch). Multiplicity is checked at load time.
func (d *Definition) DatasourceLookupOp() *Operation {
	if d == nil || d.Datasource == nil {
		return nil
	}
	for _, op := range d.Operations {
		if op.Type == OperationRead && !op.DisableView && op.DatasourceLookup {
			return op
		}
	}
	return nil
}

// DatasourceLookupID returns the name of the primary lookup attribute used to
// identify a single resource via the data source. For composite IDs, this is the
// most specific (leaf) segment; otherwise it falls back to Terraform's standard "id".
func (d *Definition) DatasourceLookupID() string {
	if n := len(d.IDAttributeComposed); n > 0 {
		return d.IDAttributeComposed[n-1]
	}
	return "id"
}

// DatasourceLookupComposedOf returns the alternative lookup attributes — sorted values
// of the datasourceLookup op's resultListLookupKeys. Nil when no datasourceLookup op exists.
func (d *Definition) DatasourceLookupComposedOf() []string {
	op := d.DatasourceLookupOp()
	if op == nil {
		return nil
	}
	out := slices.Collect(maps.Values(op.ResultListLookupKeys))
	sort.Strings(out)
	return out
}

// DatasourceLookupHas reports whether name is part of the lookup set (id or composedOf).
// Returns false on a nil Definition or when no datasourceLookup op exists.
func (d *Definition) DatasourceLookupHas(name string) bool {
	if d == nil || d.DatasourceLookupOp() == nil {
		return false
	}
	if name == d.DatasourceLookupID() {
		return true
	}
	return slices.Contains(d.DatasourceLookupComposedOf(), name)
}

// Item carries every piece of information the generator needs about a single
// schema node (root, property, array element, map value). Public-facing
// `yaml:"..."` tags mirror the keys users write in definition YAML files
// (e.g. `required`, `optional`, `computed` map into the Override* pointers
// so user values cleanly override the API-derived defaults). Internal runtime
// fields — set by createRootItem / fillOptionalFields / recalcDeep — use an
// underscore-prefixed tag (`_appearsIn`, `_required`, …) so a yaml round-trip
// can deep-copy an Item without losing computed state, without clashing with
// the user-facing keys, and without being settable from user YAML (the schema
// validator rejects unknown keys, and underscore-prefixed names are reserved
// for the generator). Only `Parent` is excluded (`yaml:"-"`) because it forms
// a cycle; deepCopyItem re-wires it after unmarshal.
type Item struct {
	Parent              *Item     `yaml:"-"`
	AppearsIn           AppearsIn `yaml:"_appearsIn,omitempty"` // In Create, Read, Update, Delete request or response
	Name                string    `yaml:"_name,omitempty"`
	IDAttribute         bool      `yaml:"_idAttribute,omitempty"` // If field is part of ID attribute or it is ID attribute
	IDAttributePosition int       `yaml:"_idAttributePosition,omitempty"`

	// Tagged fields are exposed to the Definition.yaml
	Properties map[string]*Item `yaml:"properties,omitempty"`
	Items      *Item            `yaml:"items,omitempty"` // Array item or Map item

	// Inherited from the user definition and propagated to "Items";
	// this simplifies generation logic, as both arrays and maps use "Items" similarly.
	AdditionalProperties *Item `yaml:"additionalProperties,omitempty"`

	// User-defined fields for YAML generation
	OverrideRequired   *bool `yaml:"required,omitempty"`
	OverrideComputed   *bool `yaml:"computed,omitempty"`
	OverrideOptional   *bool `yaml:"optional,omitempty"`
	OverrideSensitive  *bool `yaml:"sensitive,omitempty"`
	OverrideForceNew   *bool `yaml:"forceNew,omitempty"`
	UseStateForUnknown bool  `yaml:"useStateForUnknown,omitempty"`
	WriteOnly          bool  `yaml:"writeOnly,omitempty"`

	// FromSchemaOverride is an internal flag set during datasource.schemaOverride
	// merging on every item the overlay touched. The datasource branch of
	// IsRequired/IsOptional/IsComputed honours the merged Required/Optional/
	// Computed values for these items so user-set Required/Optional/Computed
	// (and the cascading effects of e.g. overriding an ID attribute to Optional)
	// survive without being rewritten by the entity-aware default rules.
	// Not user-settable from YAML.
	FromSchemaOverride bool `yaml:"_fromSchemaOverride,omitempty"`

	// TF Validators
	// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/validators-predefined#background
	ConflictsWith []string `yaml:"conflictsWith,omitempty"`
	ExactlyOneOf  []string `yaml:"exactlyOneOf,omitempty"`
	AtLeastOneOf  []string `yaml:"atLeastOneOf,omitempty"`
	AlsoRequires  []string `yaml:"alsoRequires,omitempty"`

	JSONName           string     `yaml:"jsonName,omitempty"`
	Type               SchemaType `yaml:"type,omitempty"`
	Description        string     `yaml:"description,omitempty"`
	DeprecationMessage string     `yaml:"deprecationMessage,omitempty"`
	Pattern            string     `yaml:"pattern,omitempty"`
	Required           bool       `yaml:"_required,omitempty"`
	Computed           bool       `yaml:"_computed,omitempty"`
	Virtual            bool       `yaml:"_virtual,omitempty"` // The field doesn't appear in API request/response. Only "id" for now.
	Nullable           bool       `yaml:"_nullable,omitempty"`
	Optional           bool       `yaml:"_optional,omitempty"`
	Sensitive          bool       `yaml:"_sensitive,omitempty"`
	ForceNew           bool       `yaml:"_forceNew,omitempty"`
	Default            any        `yaml:"default,omitempty"`
	Enum               []any      `yaml:"enum,omitempty"`
	MinLength          int        `yaml:"minLength,omitempty"`
	MaxLength          int        `yaml:"maxLength,omitempty"`
	MinItems           int        `yaml:"minItems,omitempty"`
	MaxItems           int        `yaml:"maxItems,omitempty"`
	Minimum            int        `yaml:"minimum,omitempty"`
	Maximum            int        `yaml:"maximum,omitempty"`
	Example            any        `yaml:"example,omitempty"`
}

// UniqueName generates unique name by composing all ancestor names
func (item *Item) UniqueName() string {
	p := strings.ReplaceAll(item.Path(), "/", "_")
	return firstUpper(strcase.ToGoCamel(p))
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

func (item *Item) GoType() string {
	return strings.ToLower(item.TFType())
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

func (item *Item) IsRequired(def *Definition, entity entityType) bool {
	if entity.isResource() || item.FromSchemaOverride {
		return item.Required
	}

	if item.Virtual {
		return false
	}

	// ID attributes are required in data sources, except when they participate in the lookup set.
	return item.IDAttribute && !def.DatasourceLookupHas(item.Name)
}

func (item *Item) IsOptional(def *Definition, entity entityType) bool {
	if entity.isResource() || item.FromSchemaOverride {
		return item.Optional
	}

	if item.Virtual {
		return false
	}

	return def.DatasourceLookupHas(item.Name)
}

func (item *Item) IsComputed(def *Definition, entity entityType) bool {
	if item.IsRoot() {
		return false
	}

	if entity.isResource() || item.FromSchemaOverride {
		return item.Computed
	}

	return !item.IsRequired(def, entity)
}

func (item *Item) IsReadOnly(def *Definition, entity entityType) bool {
	return !item.IsRequired(def, entity) && !item.IsOptional(def, entity)
}

func (item *Item) PropertiesByEntity(entity entityType) map[string]*Item {
	props := maps.Clone(item.Properties)
	for k, v := range item.Properties {
		if entity.IsDataSource() && v.WriteOnly {
			delete(props, k)
			for _, a := range v.AlsoRequires {
				delete(props, a)
			}
		}
	}

	return props
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
		if v.IDAttribute {
			if !v.Virtual {
				fields = append(fields, v)
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].IDAttributePosition < fields[j].IDAttributePosition
	})
	return fields
}
