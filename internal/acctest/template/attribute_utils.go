package template

import (
	"reflect"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// AttributePropertyGetter provides a common interface for accessing attribute properties
type AttributePropertyGetter struct{}

// NewAttributePropertyGetter creates a new attribute property getter
func NewAttributePropertyGetter() *AttributePropertyGetter {
	return &AttributePropertyGetter{}
}

// IsRequired checks if an attribute is required
func (g *AttributePropertyGetter) IsRequired(attr interface{}) bool {
	// Use reflection to check for the Required field
	v := reflect.ValueOf(attr)
	if !v.IsValid() {
		return false
	}

	// Handle different attribute types
	switch a := attr.(type) {
	case resourceschema.StringAttribute:
		return a.Required
	case datasourceschema.StringAttribute:
		return a.Required
	case resourceschema.BoolAttribute:
		return a.Required
	case datasourceschema.BoolAttribute:
		return a.Required
	case resourceschema.Int64Attribute:
		return a.Required
	case datasourceschema.Int64Attribute:
		return a.Required
	case resourceschema.Float64Attribute:
		return a.Required
	case datasourceschema.Float64Attribute:
		return a.Required
	case resourceschema.MapAttribute:
		return a.Required
	case datasourceschema.MapAttribute:
		return a.Required
	case resourceschema.ListAttribute:
		return a.Required
	case datasourceschema.ListAttribute:
		return a.Required
	case resourceschema.SetAttribute:
		return a.Required
	case datasourceschema.SetAttribute:
		return a.Required
	case resourceschema.ListNestedAttribute:
		return a.Required
	case datasourceschema.ListNestedAttribute:
		return a.Required
	case resourceschema.SetNestedAttribute:
		return a.Required
	case datasourceschema.SetNestedAttribute:
		return a.Required
	case resourceschema.SingleNestedAttribute:
		return a.Required
	case datasourceschema.SingleNestedAttribute:
		return a.Required
	}

	// Try to use reflection for other types
	if v.Kind() == reflect.Struct {
		requiredField := v.FieldByName("Required")
		if requiredField.IsValid() && requiredField.Kind() == reflect.Bool {
			return requiredField.Bool()
		}
	}

	return false
}

// IsOptional checks if an attribute is optional
func (g *AttributePropertyGetter) IsOptional(attr interface{}) bool {
	// Handle different attribute types
	switch a := attr.(type) {
	case resourceschema.StringAttribute:
		return a.Optional
	case datasourceschema.StringAttribute:
		return a.Optional
	case resourceschema.BoolAttribute:
		return a.Optional
	case datasourceschema.BoolAttribute:
		return a.Optional
	case resourceschema.Int64Attribute:
		return a.Optional
	case datasourceschema.Int64Attribute:
		return a.Optional
	case resourceschema.Float64Attribute:
		return a.Optional
	case datasourceschema.Float64Attribute:
		return a.Optional
	case resourceschema.MapAttribute:
		return a.Optional
	case datasourceschema.MapAttribute:
		return a.Optional
	case resourceschema.ListAttribute:
		return a.Optional
	case datasourceschema.ListAttribute:
		return a.Optional
	case resourceschema.SetAttribute:
		return a.Optional
	case datasourceschema.SetAttribute:
		return a.Optional
	case resourceschema.ListNestedAttribute:
		return a.Optional
	case datasourceschema.ListNestedAttribute:
		return a.Optional
	case resourceschema.SetNestedAttribute:
		return a.Optional
	case datasourceschema.SetNestedAttribute:
		return a.Optional
	case resourceschema.SingleNestedAttribute:
		return a.Optional
	case datasourceschema.SingleNestedAttribute:
		return a.Optional
	}

	// Try to use reflection for other types
	v := reflect.ValueOf(attr)
	if v.IsValid() && v.Kind() == reflect.Struct {
		optionalField := v.FieldByName("Optional")
		if optionalField.IsValid() && optionalField.Kind() == reflect.Bool {
			return optionalField.Bool()
		}
	}

	return false
}

// IsComputed checks if an attribute is computed
func (g *AttributePropertyGetter) IsComputed(attr interface{}) bool {
	// Handle different attribute types
	switch a := attr.(type) {
	case resourceschema.StringAttribute:
		return a.Computed
	case datasourceschema.StringAttribute:
		return a.Computed
	case resourceschema.BoolAttribute:
		return a.Computed
	case datasourceschema.BoolAttribute:
		return a.Computed
	case resourceschema.Int64Attribute:
		return a.Computed
	case datasourceschema.Int64Attribute:
		return a.Computed
	case resourceschema.Float64Attribute:
		return a.Computed
	case datasourceschema.Float64Attribute:
		return a.Computed
	case resourceschema.MapAttribute:
		return a.Computed
	case datasourceschema.MapAttribute:
		return a.Computed
	case resourceschema.ListAttribute:
		return a.Computed
	case datasourceschema.ListAttribute:
		return a.Computed
	case resourceschema.SetAttribute:
		return a.Computed
	case datasourceschema.SetAttribute:
		return a.Computed
	case resourceschema.ListNestedAttribute:
		return a.Computed
	case datasourceschema.ListNestedAttribute:
		return a.Computed
	case resourceschema.SetNestedAttribute:
		return a.Computed
	case datasourceschema.SetNestedAttribute:
		return a.Computed
	case resourceschema.SingleNestedAttribute:
		return a.Computed
	case datasourceschema.SingleNestedAttribute:
		return a.Computed
	}

	// Try to use reflection for other types
	v := reflect.ValueOf(attr)
	if v.IsValid() && v.Kind() == reflect.Struct {
		computedField := v.FieldByName("Computed")
		if computedField.IsValid() && computedField.Kind() == reflect.Bool {
			return computedField.Bool()
		}
	}

	return false
}
