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
	"strings"
)

// ResourceKind represents the type of terraform configuration item
type ResourceKind string

// Possible values for ResourceKind
const (
	ResourceKindResource   ResourceKind = "resource"
	ResourceKindDataSource ResourceKind = "data"
)

// String returns the string representation of ResourceKind
func (k ResourceKind) String() string {
	switch k {
	case ResourceKindResource:
		return "resource"
	case ResourceKindDataSource:
		return "data"
	default:
		return "unknown"
	}
}

// TemplateGenerator is the base interface for all template generators
type TemplateGenerator interface {
	// GenerateTemplate generates a template for the given schema, resource type, and kind
	GenerateTemplate(schema interface{}, resourceType string, kind ResourceKind) string
}

// SchemaFieldExtractor is an interface for extracting fields from different schema types
type SchemaFieldExtractor interface {
	// ExtractFields extracts fields from a schema
	ExtractFields(schema interface{}) ([]TemplateField, error)
}

// TimeoutsConfig defines which timeouts are configured for a resource
type TimeoutsConfig struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
}

// TemplatePath represents a field access path in a Go template
type TemplatePath struct {
	// For the following example:
	//
	// resource "example" "foo" {
	//   top {
	//     nested {
	//       field = "value"
	//     }
	//   }
	// }
	//
	// `components` would be: ["top", "nested", "field"] and
	// `isCollection` would be: [true, true, false]

	components   []string
	isCollection []bool
}

// NewTemplatePath creates a new template path with a top-level field
func NewTemplatePath(fieldName string, isCollection bool) TemplatePath {
	return TemplatePath{
		components:   []string{fieldName},
		isCollection: []bool{isCollection},
	}
}

// AppendField adds a new field to the path
func (p TemplatePath) AppendField(fieldName string, isCollection bool) TemplatePath {
	p.components = append(p.components, fieldName)
	p.isCollection = append(p.isCollection, isCollection)
	return p
}

// Expression returns the template expression for accessing this path
func (p TemplatePath) Expression() string {
	if len(p.components) == 0 {
		return ""
	}

	// Start with the root
	expr := "." + p.components[0]

	// Build the expression with string-indexed paths
	for i := 1; i < len(p.components); i++ {
		if p.isCollection[i-1] {
			// For collections, use numeric index 0 followed by string index
			expr = fmt.Sprintf("(index %s 0 %q)", expr, p.components[i])
		} else {
			// For regular fields, use string index
			expr = fmt.Sprintf("(index %s %q)", expr, p.components[i])
		}
	}

	return expr
}

// CommonTemplateRenderer provides shared template rendering logic
// for both SDK and Framework template generators
type CommonTemplateRenderer struct{}

// RenderTimeouts renders a timeouts block with standard structure
func (r *CommonTemplateRenderer) RenderTimeouts(builder *strings.Builder, indent int, config TimeoutsConfig) {
	indentStr := strings.Repeat("  ", indent)

	// Add top-level conditional for the entire timeouts block
	fmt.Fprintf(builder, "%s{{- if .timeouts }}\n", indentStr)
	fmt.Fprintf(builder, "%stimeouts {\n", indentStr)

	// Only render the timeouts that are configured in the schema
	if config.Create {
		fmt.Fprintf(builder, "%s  {{- if .timeouts.create }}\n", indentStr)
		fmt.Fprintf(builder, "%s  create = {{ renderValue .timeouts.create }}\n", indentStr)
		fmt.Fprintf(builder, "%s  {{- end }}\n", indentStr)
	}

	if config.Read {
		fmt.Fprintf(builder, "%s  {{- if .timeouts.read }}\n", indentStr)
		fmt.Fprintf(builder, "%s  read = {{ renderValue .timeouts.read }}\n", indentStr)
		fmt.Fprintf(builder, "%s  {{- end }}\n", indentStr)
	}

	if config.Update {
		fmt.Fprintf(builder, "%s  {{- if .timeouts.update }}\n", indentStr)
		fmt.Fprintf(builder, "%s  update = {{ renderValue .timeouts.update }}\n", indentStr)
		fmt.Fprintf(builder, "%s  {{- end }}\n", indentStr)
	}

	if config.Delete {
		fmt.Fprintf(builder, "%s  {{- if .timeouts.delete }}\n", indentStr)
		fmt.Fprintf(builder, "%s  delete = {{ renderValue .timeouts.delete }}\n", indentStr)
		fmt.Fprintf(builder, "%s  {{- end }}\n", indentStr)
	}

	fmt.Fprintf(builder, "%s}\n", indentStr)
	fmt.Fprintf(builder, "%s{{- end }}\n", indentStr)
}

// RenderDependsOn renders a depends_on attribute with standard format
func (r *CommonTemplateRenderer) RenderDependsOn(builder *strings.Builder, indent int) {
	indentStr := strings.Repeat("  ", indent)

	fmt.Fprintf(builder, "%s{{- if .depends_on }}\n", indentStr)
	fmt.Fprintf(builder, "%sdepends_on = [%s", indentStr, "")
	fmt.Fprintf(builder, "%s", "{{- range $i, $dep := .depends_on }}{{if $i}}, {{end}}{{ renderValue $dep }}{{- end }}]\n")
	fmt.Fprintf(builder, "%s{{- end }}\n", indentStr)
}

// RenderField renders a template field based on its properties
func (r *CommonTemplateRenderer) RenderField(builder *strings.Builder, field TemplateField, indent int, parentPath TemplatePath) {
	// Create a path for this field
	var path TemplatePath
	if parentPath.components == nil {
		// This is a top-level field
		path = NewTemplatePath(field.Name, field.IsCollection)
	} else {
		// This is a nested field
		path = parentPath.AppendField(field.Name, field.IsCollection)
	}

	if field.IsObject && len(field.NestedFields) > 0 {
		r.renderBlock(builder, field, path, indent)
	} else if field.IsCollection && !field.IsObject {
		r.renderCollection(builder, field, path, indent)
	} else if field.IsMap {
		r.renderMap(builder, field, path, indent)
	} else if field.FieldType == FieldTypeBool {
		r.renderBool(builder, field, path, indent)
	} else if field.FieldType == FieldTypeNumber {
		r.renderSimple(builder, field, path, indent)
	} else {
		r.renderSimple(builder, field, path, indent)
	}
}

// renderFieldWithContent is a helper that handles the common pattern of optional/required fields
func (r *CommonTemplateRenderer) renderFieldWithContent(builder *strings.Builder, field TemplateField, path TemplatePath, indent int,
	renderFunc func(*strings.Builder, TemplateField, string, int),
) {
	indentStr := strings.Repeat("  ", indent)
	pathExpr := path.Expression()

	if !field.Required || field.Optional {
		// Optional fields need an existence check
		fmt.Fprintf(builder, "%s{{- if %s }}\n", indentStr, pathExpr)
	}

	// Render the specific content for this field type
	renderFunc(builder, field, pathExpr, indent)

	if !field.Required || field.Optional {
		// Close the conditional for optional fields
		fmt.Fprintf(builder, "%s{{- end }}\n", indentStr)
	}
}

// renderBlock handles a block with nested fields
func (r *CommonTemplateRenderer) renderBlock(builder *strings.Builder, field TemplateField, path TemplatePath, indent int) {
	r.renderFieldWithContent(builder, field, path, indent, func(b *strings.Builder, field TemplateField, _ string, indent int) {
		indentStr := strings.Repeat("  ", indent)
		fmt.Fprintf(b, "%s%s {\n", indentStr, field.Name)

		for _, nestedField := range field.NestedFields {
			// Pass the current path as parent for nested fields
			r.RenderField(b, nestedField, indent+1, path)
		}

		fmt.Fprintf(b, "%s}\n", indentStr)
	})
}

// renderSimple handles simple fields (strings, etc.)
func (r *CommonTemplateRenderer) renderSimple(builder *strings.Builder, field TemplateField, path TemplatePath, indent int) {
	r.renderFieldWithContent(builder, field, path, indent, func(b *strings.Builder, field TemplateField, pathExpr string, indent int) {
		indentStr := strings.Repeat("  ", indent)
		if field.Required {
			// Add required wrapper for required fields
			fmt.Fprintf(b, "%s%s = {{ renderValue (required %s) }}\n", indentStr, field.Name, path.Expression())
		} else {
			fmt.Fprintf(b, "%s%s = {{ renderValue %s }}\n", indentStr, field.Name, pathExpr)
		}
	})
}

// renderBool handles boolean fields with special null handling
func (r *CommonTemplateRenderer) renderBool(builder *strings.Builder, field TemplateField, path TemplatePath, indent int) {
	indentStr := strings.Repeat("  ", indent)
	pathExpr := path.Expression()

	if field.Required {
		// Just use the standard path expression for required booleans
		fmt.Fprintf(builder, "%s%s = {{ %s }}\n", indentStr, field.Name, pathExpr)
	} else {
		// Build condition with the prefix "ne ... nil" to check existence, not value
		fmt.Fprintf(builder, "%s{{- if ne %s nil }}\n", indentStr, pathExpr)
		fmt.Fprintf(builder, "%s%s = {{ %s }}\n", indentStr, field.Name, pathExpr)
		fmt.Fprintf(builder, "%s{{- end }}\n", indentStr)
	}
}

// renderMap handles map fields
func (r *CommonTemplateRenderer) renderMap(builder *strings.Builder, field TemplateField, path TemplatePath, indent int) {
	r.renderFieldWithContent(builder, field, path, indent, func(b *strings.Builder, field TemplateField, pathExpr string, indent int) {
		indentStr := strings.Repeat("  ", indent)
		fmt.Fprintf(b, "%s%s = {\n", indentStr, field.Name)
		fmt.Fprintf(b, "%s  {{- range $k, $v := %s }}\n", indentStr, pathExpr)
		fmt.Fprintf(b, "%s  {{ renderValue $k }} = {{ renderValue $v }}\n", indentStr)
		fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		fmt.Fprintf(b, "%s}\n", indentStr)
	})
}

// renderCollection handles list/set fields of primitive values
func (r *CommonTemplateRenderer) renderCollection(builder *strings.Builder, field TemplateField, path TemplatePath, indent int) {
	r.renderFieldWithContent(builder, field, path, indent, func(b *strings.Builder, field TemplateField, pathExpr string, indent int) {
		indentStr := strings.Repeat("  ", indent)
		fmt.Fprintf(b, "%s%s = [\n", indentStr, field.Name)
		fmt.Fprintf(b, "%s  {{- range $idx, $item := %s }}\n", indentStr, pathExpr)
		fmt.Fprintf(b, "%s  {{ renderValue $item }},\n", indentStr)
		fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		fmt.Fprintf(b, "%s]\n", indentStr)
	})
}
