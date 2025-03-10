package template

import (
	"fmt"
	"reflect"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FieldType represents the type of a field in a schema
type FieldType int

const (
	// https://developer.hashicorp.com/terraform/language/expressions/types
	FieldTypeUnknown FieldType = iota
	FieldTypeString
	FieldTypeNumber
	FieldTypeBool
	FieldTypeMap
	FieldTypeCollection
	FieldTypeObject
)

// TemplateField represents a field in a schema with its properties
type TemplateField struct {
	Name         string
	Required     bool
	Optional     bool
	Computed     bool
	FieldType    FieldType
	NestedFields []TemplateField
	IsObject     bool
	IsCollection bool
	IsMap        bool
}

// FrameworkFieldExtractor extracts fields from Framework schema
type FrameworkFieldExtractor struct{}

// NewFrameworkFieldExtractor creates a new extractor for Framework schemas
func NewFrameworkFieldExtractor() *FrameworkFieldExtractor {
	return &FrameworkFieldExtractor{}
}

var _ SchemaFieldExtractor = &FrameworkFieldExtractor{}

// ExtractFields implements SchemaFieldExtractor for Framework schemas
func (e *FrameworkFieldExtractor) ExtractFields(schema interface{}) ([]TemplateField, error) {
	switch s := schema.(type) {
	case resourceschema.Schema:
		return e.extractFields(s.Attributes), nil
	case datasourceschema.Schema:
		return e.extractFields(s.Attributes), nil
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", schema)
	}
}

// extractFields extracts fields from either resource or datasource attributes
func (e *FrameworkFieldExtractor) extractFields(attributes interface{}) []TemplateField {
	fields := make([]TemplateField, 0)

	// Handle the different attribute map types with a type switch
	switch attrs := attributes.(type) {
	case map[string]resourceschema.Attribute:
		for name, attr := range attrs {
			e.processField(name, attr, &fields)
		}
	case map[string]datasourceschema.Attribute:
		for name, attr := range attrs {
			e.processField(name, attr, &fields)
		}
	}

	return fields
}

// processField processes a single field and adds it to the fields list if not skipped
func (e *FrameworkFieldExtractor) processField(name string, attr interface{}, fields *[]TemplateField) {
	// Create field with common properties
	field := TemplateField{
		Name:      name,
		Required:  e.isRequired(attr),
		Optional:  e.isOptional(attr),
		Computed:  e.isComputed(attr),
		FieldType: FieldTypeUnknown, // Default to unknown
	}

	switch a := attr.(type) {
	case resourceschema.BoolAttribute, datasourceschema.BoolAttribute:
		field.FieldType = FieldTypeBool
	case resourceschema.Int64Attribute, resourceschema.Float64Attribute, resourceschema.NumberAttribute,
		datasourceschema.Int64Attribute, datasourceschema.Float64Attribute, datasourceschema.NumberAttribute:
		field.FieldType = FieldTypeNumber
	case resourceschema.MapAttribute, datasourceschema.MapAttribute:
		field.IsMap = true
		field.FieldType = FieldTypeMap
	case resourceschema.ListAttribute, resourceschema.SetAttribute,
		datasourceschema.ListAttribute, datasourceschema.SetAttribute:
		field.IsCollection = true
		field.FieldType = FieldTypeCollection
	case resourceschema.ListNestedAttribute, resourceschema.SetNestedAttribute:
		field.IsCollection = true
		field.IsObject = true
		field.FieldType = FieldTypeObject
		// Type assertion is needed to access the NestedObject.Attributes
		switch nested := a.(type) {
		case resourceschema.ListNestedAttribute:
			field.NestedFields = e.extractFields(nested.NestedObject.Attributes)
		case resourceschema.SetNestedAttribute:
			field.NestedFields = e.extractFields(nested.NestedObject.Attributes)
		}
	case datasourceschema.ListNestedAttribute, datasourceschema.SetNestedAttribute:
		field.IsCollection = true
		field.IsObject = true
		field.FieldType = FieldTypeObject
		// Type assertion is needed to access the NestedObject.Attributes
		switch nested := a.(type) {
		case datasourceschema.ListNestedAttribute:
			field.NestedFields = e.extractFields(nested.NestedObject.Attributes)
		case datasourceschema.SetNestedAttribute:
			field.NestedFields = e.extractFields(nested.NestedObject.Attributes)
		}
	case resourceschema.SingleNestedAttribute:
		field.IsObject = true
		field.FieldType = FieldTypeObject
		field.NestedFields = e.extractFields(a.Attributes)
	case datasourceschema.SingleNestedAttribute:
		field.IsObject = true
		field.FieldType = FieldTypeObject
		field.NestedFields = e.extractFields(a.Attributes)
	}

	*fields = append(*fields, field)
}

// getFieldBool checks if a boolean field is set on an attribute
func (e *FrameworkFieldExtractor) getFieldBool(attr interface{}, fieldName string) bool {
	v := reflect.ValueOf(attr)
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName(fieldName); f.IsValid() && f.Kind() == reflect.Bool { // nosemgrep
			return f.Bool()
		}
	}
	return false
}

// isRequired checks if an attribute is required
func (e *FrameworkFieldExtractor) isRequired(attr interface{}) bool {
	return e.getFieldBool(attr, "Required")
}

// isOptional checks if an attribute is optional
func (e *FrameworkFieldExtractor) isOptional(attr interface{}) bool {
	return e.getFieldBool(attr, "Optional")
}

// isComputed checks if an attribute is computed
func (e *FrameworkFieldExtractor) isComputed(attr interface{}) bool {
	return e.getFieldBool(attr, "Computed")
}

// SDKFieldExtractor extracts fields from SDK schema
type SDKFieldExtractor struct{}

// NewSDKFieldExtractor creates a new extractor for SDK schemas
func NewSDKFieldExtractor() *SDKFieldExtractor {
	return &SDKFieldExtractor{}
}

var _ SchemaFieldExtractor = &SDKFieldExtractor{}

// ExtractFields implements SchemaFieldExtractor for SDK schemas
func (e *SDKFieldExtractor) ExtractFields(schema interface{}) ([]TemplateField, error) {
	resource, ok := schema.(*sdkschema.Resource)
	if !ok {
		return nil, fmt.Errorf("schema is not a *schema.Resource: %T", schema)
	}

	return e.processSchema(resource.Schema), nil
}

// processSchema extracts fields from an SDK schema
func (e *SDKFieldExtractor) processSchema(schema map[string]*sdkschema.Schema) []TemplateField {
	fields := make([]TemplateField, 0, len(schema))

	for name, sch := range schema {
		// Skip computed-only fields (that aren't required or optional)
		if sch.Computed && !sch.Optional && !sch.Required {
			continue
		}

		field := TemplateField{
			Name:      name,
			Required:  sch.Required,
			Optional:  sch.Optional,
			Computed:  sch.Computed,
			FieldType: FieldTypeUnknown, // Default to unknown
		}

		// Handle different schema types
		switch sch.Type {
		case sdkschema.TypeBool:
			field.FieldType = FieldTypeBool
		case sdkschema.TypeInt, sdkschema.TypeFloat:
			field.FieldType = FieldTypeNumber
		case sdkschema.TypeMap:
			field.IsMap = true
			field.FieldType = FieldTypeMap
		case sdkschema.TypeList, sdkschema.TypeSet:
			field.IsCollection = true
			field.FieldType = FieldTypeCollection

			// Check if Elem is a resource (nested block)
			if res, ok := sch.Elem.(*sdkschema.Resource); ok {
				field.IsObject = true
				field.FieldType = FieldTypeObject
				// Recursively process nested schema
				field.NestedFields = e.processSchema(res.Schema)
			}
		}

		fields = append(fields, field)
	}

	return fields
}
