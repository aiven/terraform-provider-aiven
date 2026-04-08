package adapter

type SchemaType string

const (
	SchemaTypeString SchemaType = "string"
	SchemaTypeFloat  SchemaType = "float"
	SchemaTypeInt    SchemaType = "int"
	SchemaTypeBool   SchemaType = "bool"
	SchemaTypeList   SchemaType = "list"
	SchemaTypeSet    SchemaType = "set"
	SchemaTypeMap    SchemaType = "map"
	SchemaTypeObject SchemaType = "object"
)

// Schema simplified schema similar to OpenAPI spec.
// Used internally to perform data normalization and validation,
// because TF schemas have different types for resource and datasource,
// they use blocks and attributes.
type Schema struct {
	Type       SchemaType
	Properties map[string]*Schema
	Items      *Schema

	// Computed fields are allowed to be read from the state.
	Computed bool

	// In SDKv2, an object schema is represented as a list containing a single object.
	// The path syntax reflects this, using indices such as "foo.0.bar", where "bar" is a property of the single object, accessed via index 0.
	// This flag distinguishes between a direct object and a list containing one object, which is important for marshaling.
	IsObject bool

	// If true, the field value cannot be zero (i.e., it has a minimum, minLength, or minItems > 0).
	// In SDKv2, the state may contain empty values that should not be sent to the API.
	// Because we can't distinguish between zero values and missing values in the SDKv2 state,
	// this flag allows such cases to be ignored.
	ZeroNotAllowed bool
}

func isScalar(kind SchemaType) bool {
	switch kind {
	case SchemaTypeString, SchemaTypeInt, SchemaTypeFloat, SchemaTypeBool:
		return true
	}
	return false
}
