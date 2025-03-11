// Package template provides a framework for generating Terraform configuration templates
// from schema definitions across different Terraform development frameworks.
//
// This package enables automated template generation for both Terraform SDK v2 and
// Plugin Framework schemas through a unified interface. The core workflow is:
//
//  1. Schema Analysis: Extract fields and their properties (required, optional, computed)
//     from different schema types (SDK v2 Resources or Plugin Framework Schemas)
//
//  2. Field Processing: Determine field characteristics (type, nested structure, etc.)
//     and organize them into a unified TemplateField representation
//
//  3. Template Generation: Convert the structured field data into properly formatted
//     Terraform configuration templates with appropriate conditionals and formatting
//
// The package uses a modular design with discrete interfaces for each part of the process:
// - TemplateGenerator: Main entry point for template generation
// - SchemaFieldExtractor: Extracts fields from different schema types
//
// By separating these concerns, the package supports different schema types through
// specialized implementations while maintaining a consistent template generation pipeline.
package template

import (
	"fmt"
)

// Template represents a single Terraform configuration template
type Template struct {
	Name     string
	Template string
}

// Value represents a value that can either be a literal string or a reference
type Value struct {
	Value     string
	IsLiteral bool
}

// Literal creates a new literal value
func Literal(v string) Value {
	return Value{Value: v, IsLiteral: true}
}

// Reference creates a new reference value
func Reference(v string) Value {
	return Value{Value: v, IsLiteral: false}
}

// String returns the properly formatted value based on whether it's a literal or reference
func (v Value) String() string {
	if v.IsLiteral {
		return fmt.Sprintf("%q", v.Value)
	}
	return v.Value
}

// MarshalText implements encoding.TextMarshaler
func (v Value) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
