package template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// FrameworkSchemaTemplateGenerator interface for generating templates from Framework schemas
type FrameworkSchemaTemplateGenerator interface {
	TemplateGenerator
	GenerateFrameworkTemplate(schema interface{}, resourceType string, kind ResourceKind) string
}

// FrameworkTextTemplateGenerator creates Go templates from Terraform Framework resource schemas
type FrameworkTextTemplateGenerator struct {
	propertyGetter *AttributePropertyGetter
}

// NewFrameworkSchemaTemplateGenerator creates a new template generator for Framework resources
func NewFrameworkSchemaTemplateGenerator() *FrameworkTextTemplateGenerator {
	return &FrameworkTextTemplateGenerator{
		propertyGetter: NewAttributePropertyGetter(),
	}
}

// GenerateTemplate implements the TemplateGenerator interface
func (g *FrameworkTextTemplateGenerator) GenerateTemplate(schema interface{}, resourceType string, kind ResourceKind) string {
	// For Framework schemas, we expect either resourceschema.Schema or datasourceschema.Schema
	switch s := schema.(type) {
	case resourceschema.Schema:
		return g.GenerateFrameworkTemplate(s, resourceType, kind)
	case datasourceschema.Schema:
		return g.GenerateFrameworkTemplate(s, resourceType, kind)
	default:
		// If not a recognized schema type, return a basic template with an error comment
		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")
		b.WriteString("  # Error: Framework generator received non-Framework schema\n")
		b.WriteString("}")
		return b.String()
	}
}

// GenerateFrameworkTemplate generates a Go template string that resembles Terraform HCL configuration
// for a Framework resource or data source.
func (g *FrameworkTextTemplateGenerator) GenerateFrameworkTemplate(schema interface{}, resourceType string, kind ResourceKind) string {
	var b strings.Builder

	// Header differs based on kind
	_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")

	switch s := schema.(type) {
	case resourceschema.Schema:
		g.generateFields(&b, s.Attributes, 1)
	case datasourceschema.Schema:
		g.generateFields(&b, s.Attributes, 1)
	default:
		// If schema type is not recognized, return a basic template
		b.WriteString("  # Schema type not recognized\n")
	}

	b.WriteString("}")

	return b.String()
}

// generateFields generates template fields from Framework schema attributes
func (g *FrameworkTextTemplateGenerator) generateFields(b *strings.Builder, attributes interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Collect fields by type
	var (
		required []string
		optional []string
		computed []string
	)

	// Convert attributes to a map we can iterate over
	attrMap := make(map[string]interface{})

	switch attrs := attributes.(type) {
	case map[string]resourceschema.Attribute:
		for k, v := range attrs {
			attrMap[k] = v
		}
	case map[string]schema.Attribute:
		for k, v := range attrs {
			attrMap[k] = v
		}
	default:
		// If we can't determine the type, return early
		return
	}

	for k, field := range attrMap {
		// Skip resource_name as it's handled in the template header
		if k == "resource_name" {
			continue
		}

		// Determine if the field is required, optional, or computed
		isRequired := g.isRequired(field)
		isComputed := g.isComputed(field)
		isOptional := g.isOptional(field)

		if isRequired {
			required = append(required, k)
		} else if isOptional && !isComputed {
			optional = append(optional, k)
		} else if isComputed && !isOptional {
			computed = append(computed, k)
		} else {
			// Both optional and computed, treat as optional
			optional = append(optional, k)
		}
	}

	// Sort all field groups for consistent ordering
	sort.Strings(required)
	sort.Strings(optional)
	sort.Strings(computed)

	// Process fields in specified order
	for _, field := range required {
		g.generateField(b, field, attrMap[field], indentStr, true)
	}

	for _, field := range optional {
		g.generateField(b, field, attrMap[field], indentStr, false)
	}

	// Computed-only fields are not included in the template
}

// generateField generates a template field for a Framework schema attribute
func (g *FrameworkTextTemplateGenerator) generateField(b *strings.Builder, field string, attr interface{}, indent string, required bool) {
	switch a := attr.(type) {
	case resourceschema.StringAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.StringAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case resourceschema.BoolAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.BoolAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case resourceschema.Int64Attribute, resourceschema.Float64Attribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.Int64Attribute, schema.Float64Attribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case resourceschema.MapAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {\n", indent, field)
			fmt.Fprintf(b, "%s  {{- range $k, $v := (required .%s) }}\n", indent, field)
			fmt.Fprintf(b, "%s  {{ renderValue $k }} = {{ renderValue $v }}\n", indent)
			fmt.Fprintf(b, "%s  {{- end }}\n", indent)
			fmt.Fprintf(b, "%s}\n", indent)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {\n", indent, field)
			fmt.Fprintf(b, "%s  {{- range $k, $v := .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s  {{ renderValue $k }} = {{ renderValue $v }}\n", indent)
			fmt.Fprintf(b, "%s  {{- end }}\n", indent)
			fmt.Fprintf(b, "%s}\n", indent)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.MapAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {\n", indent, field)
			fmt.Fprintf(b, "%s  {{- range $k, $v := (required .%s) }}\n", indent, field)
			fmt.Fprintf(b, "%s  {{ renderValue $k }} = {{ renderValue $v }}\n", indent)
			fmt.Fprintf(b, "%s  {{- end }}\n", indent)
			fmt.Fprintf(b, "%s}\n", indent)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {\n", indent, field)
			fmt.Fprintf(b, "%s  {{- range $k, $v := .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s  {{ renderValue $k }} = {{ renderValue $v }}\n", indent)
			fmt.Fprintf(b, "%s  {{- end }}\n", indent)
			fmt.Fprintf(b, "%s}\n", indent)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case resourceschema.ListAttribute, resourceschema.SetAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.ListAttribute, schema.SetAttribute:
		if required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case resourceschema.ListNestedAttribute:
		g.generateNestedBlock(b, field, a.NestedObject.Attributes, indent, required)
	case resourceschema.SetNestedAttribute:
		g.generateNestedBlock(b, field, a.NestedObject.Attributes, indent, required)
	case schema.ListNestedAttribute:
		g.generateNestedBlock(b, field, a.NestedObject.Attributes, indent, required)
	case schema.SetNestedAttribute:
		g.generateNestedBlock(b, field, a.NestedObject.Attributes, indent, required)

	case resourceschema.SingleNestedAttribute:
		g.generateSingleNestedBlock(b, field, a.Attributes, indent, required)

	case schema.SingleNestedAttribute:
		g.generateSingleNestedBlock(b, field, a.Attributes, indent, required)

	default:
		// For unknown attribute types, use a generic approach
		if required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}
	}
}

// generateNestedBlock generates a template for a nested block attribute
func (g *FrameworkTextTemplateGenerator) generateNestedBlock(b *strings.Builder, field string, attributes interface{}, indent string, required bool) {
	if required {
		fmt.Fprintf(b, "%s%s {\n", indent, field)
		indentLevel := len(indent)/2 + 1
		g.generateFields(b, attributes, indentLevel)
		fmt.Fprintf(b, "%s}\n", indent)
	} else {
		fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
		fmt.Fprintf(b, "%s%s {\n", indent, field)
		indentLevel := len(indent)/2 + 1
		g.generateFields(b, attributes, indentLevel)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%s{{- end }}\n", indent)
	}
}

// generateSingleNestedBlock generates a template for a single nested block attribute
func (g *FrameworkTextTemplateGenerator) generateSingleNestedBlock(b *strings.Builder, field string, attributes interface{}, indent string, required bool) {
	if required {
		fmt.Fprintf(b, "%s%s {\n", indent, field)
		indentLevel := len(indent)/2 + 1
		g.generateFields(b, attributes, indentLevel)
		fmt.Fprintf(b, "%s}\n", indent)
	} else {
		fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
		fmt.Fprintf(b, "%s%s {\n", indent, field)
		indentLevel := len(indent)/2 + 1
		g.generateFields(b, attributes, indentLevel)
		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%s{{- end }}\n", indent)
	}
}

// Helper methods to determine attribute properties

// isRequired checks if an attribute is required
func (g *FrameworkTextTemplateGenerator) isRequired(attr interface{}) bool {
	return g.propertyGetter.IsRequired(attr)
}

// isOptional checks if an attribute is optional
func (g *FrameworkTextTemplateGenerator) isOptional(attr interface{}) bool {
	return g.propertyGetter.IsOptional(attr)
}

// isComputed checks if an attribute is computed
func (g *FrameworkTextTemplateGenerator) isComputed(attr interface{}) bool {
	return g.propertyGetter.IsComputed(attr)
}
