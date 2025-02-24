package template

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceKind represents the type of terraform configuration item
type ResourceKind int

const (
	ResourceKindResource ResourceKind = iota
	ResourceKindDataSource
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

type SchemaTemplateGenerator interface {
	GenerateTemplate(r *schema.Resource, resourceType string, kind ResourceKind) string
}

// TextTemplateGenerator creates Go templates from Terraform resource schemas for testing purposes.
// It analyzes a resource's schema definition and generates a template that mirrors the structure
// of a Terraform configuration, but with template variables for dynamic values.
type TextTemplateGenerator struct {
}

// NewSchemaTemplateGenerator creates a new template generator for a specific resource type
func NewSchemaTemplateGenerator() *TextTemplateGenerator {
	return &TextTemplateGenerator{}
}

// GenerateTemplate generates a Go template string that resembles Terraform HCL configuration.
// The generated template is not actual Terraform code, but rather a template that can be rendered
// with different values to produce valid Terraform configurations for testing purposes.
//
// The template variables can then be populated using a template.Config to generate
// different variations of the resource configuration for testing.
func (g *TextTemplateGenerator) GenerateTemplate(r *schema.Resource, resourceType string, kind ResourceKind) string {
	var b strings.Builder

	// Header differs based on kind
	_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")

	g.generateFields(&b, r.Schema, 1)
	g.generateTimeouts(&b, r.Timeouts, 1)
	if _, hasDependsOn := r.Schema["depends_on"]; hasDependsOn {
		g.generateDependsOn(&b, 1)
	}

	b.WriteString("}")

	return b.String()
}

func (g *TextTemplateGenerator) generateFields(b *strings.Builder, s map[string]*schema.Schema, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Collect fields by type
	var (
		required []string
		optional []string
		lists    []string
		maps     []string
	)

	for k, field := range s {
		if field.Computed && !field.Optional && !field.Required {
			continue
		}

		switch field.Type {
		case schema.TypeList, schema.TypeSet:
			lists = append(lists, k)
		case schema.TypeMap:
			maps = append(maps, k)
		default:
			if field.Required {
				required = append(required, k)
			} else {
				optional = append(optional, k)
			}
		}
	}

	// Sort all field groups for consistent ordering
	sort.Strings(required)
	sort.Strings(optional)
	sort.Strings(lists)
	sort.Strings(maps)

	// Process fields in specified order
	for _, field := range required {
		g.generateField(b, field, s[field], indentStr)
	}

	for _, field := range optional {
		g.generateField(b, field, s[field], indentStr)
	}

	for _, field := range lists {
		g.generateField(b, field, s[field], indentStr)
	}

	for _, field := range maps {
		g.generateField(b, field, s[field], indentStr)
	}
}

func (g *TextTemplateGenerator) generateField(b *strings.Builder, field string, schemaField *schema.Schema, indent string) {
	switch schemaField.Type {
	case schema.TypeList, schema.TypeSet:
		if schemaField.Optional || schemaField.Required {
			g.generateNestedBlock(b, field, schemaField, indent, schemaField.Required)
		}
	case schema.TypeString:
		if schemaField.Required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.TypeBool:
		if schemaField.Required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			// For booleans, we want to render false values too
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.TypeInt, schema.TypeFloat:
		if schemaField.Required {
			fmt.Fprintf(b, "%s%s = {{ required .%s }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if ne .%s nil }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}

	case schema.TypeMap:
		if schemaField.Required {
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

	default:
		if schemaField.Required {
			fmt.Fprintf(b, "%s%s = {{ renderValue (required .%s) }}\n", indent, field, field)
		} else {
			fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
			fmt.Fprintf(b, "%s%s = {{ renderValue .%s }}\n", indent, field, field)
			fmt.Fprintf(b, "%s{{- end }}\n", indent)
		}
	}
}

// generateTimeouts AddTemplate timeouts block if timeouts are configured
func (g *TextTemplateGenerator) generateTimeouts(b *strings.Builder, timeouts *schema.ResourceTimeout, indent int) {
	if timeouts == nil {
		return
	}

	indentStr := strings.Repeat("  ", indent)
	hasTimeouts := false

	// Check if any timeouts are configured in schema
	if timeouts.Create != nil || timeouts.Read != nil || timeouts.Update != nil || timeouts.Delete != nil {
		hasTimeouts = true
	}

	// Only generate the timeouts block if there are timeouts configured
	if hasTimeouts {
		// AddTemplate top-level conditional for the entire timeouts block
		fmt.Fprintf(b, "%s{{- if .timeouts }}\n", indentStr)
		fmt.Fprintf(b, "%stimeouts {\n", indentStr)

		if timeouts.Create != nil {
			fmt.Fprintf(b, "%s  {{- if .timeouts.create }}\n", indentStr)
			fmt.Fprintf(b, "%s  create = {{ renderValue .timeouts.create }}\n", indentStr)
			fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		}
		if timeouts.Read != nil {
			fmt.Fprintf(b, "%s  {{- if .timeouts.read }}\n", indentStr)
			fmt.Fprintf(b, "%s  read = {{ renderValue .timeouts.read }}\n", indentStr)
			fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		}
		if timeouts.Update != nil {
			fmt.Fprintf(b, "%s  {{- if .timeouts.update }}\n", indentStr)
			fmt.Fprintf(b, "%s  update = {{ renderValue .timeouts.update }}\n", indentStr)
			fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		}
		if timeouts.Delete != nil {
			fmt.Fprintf(b, "%s  {{- if .timeouts.delete }}\n", indentStr)
			fmt.Fprintf(b, "%s  delete = {{ renderValue .timeouts.delete }}\n", indentStr)
			fmt.Fprintf(b, "%s  {{- end }}\n", indentStr)
		}

		fmt.Fprintf(b, "%s}\n", indentStr)
		fmt.Fprintf(b, "%s{{- end }}\n", indentStr)
	}
}

// generateDependsOn AddTemplate depends_on block if dependencies are specified
func (g *TextTemplateGenerator) generateDependsOn(b *strings.Builder, indent int) {
	indentStr := strings.Repeat("  ", indent)

	fmt.Fprintf(b, "%s{{- if .depends_on }}\n", indentStr)
	fmt.Fprintf(b, "%sdepends_on = [%s", indentStr, "")
	fmt.Fprintf(b, "%s", "{{- range $i, $dep := .depends_on }}{{if $i}}, {{end}}{{ renderValue $dep }}{{- end }}]\n")
	fmt.Fprintf(b, "%s{{- end }}\n", indentStr)
}

func (g *TextTemplateGenerator) generateNestedBlock(b *strings.Builder, field string, schemaField *schema.Schema, indent string, _ bool) {
	elem := schemaField.Elem
	switch e := elem.(type) {
	case *schema.Resource:
		fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
		fmt.Fprintf(b, "%s%s {\n", indent, field)

		nestedIndent := indent + "  "

		// Filter and sort fields
		var fields []string
		for k, v := range e.Schema {
			// Include field if it's either:
			// 1. Not computed, or
			// 2. Both computed and optional (but skip purely computed fields)
			if !v.Computed || v.Optional {
				fields = append(fields, k)
			}
		}
		sort.Strings(fields)

		for _, nestedField := range fields {
			nestedSchema := e.Schema[nestedField]

			switch nestedSchema.Type {
			case schema.TypeList, schema.TypeSet:
				if res, ok := nestedSchema.Elem.(*schema.Resource); ok {
					g.generateListBlock(b, field, nestedField, res, nestedIndent)
				} else if elemSchema, ok := nestedSchema.Elem.(*schema.Schema); ok {
					g.generatePrimitiveList(b, field, nestedField, elemSchema, nestedIndent)
				}
			case schema.TypeBool:
				fmt.Fprintf(b, "%s{{- if ne (index .%s 0 \"%s\") nil }}\n", nestedIndent, field, nestedField)
				fmt.Fprintf(b, "%s%s = {{ index .%s 0 \"%s\" }}\n", nestedIndent, nestedField, field, nestedField)
				fmt.Fprintf(b, "%s{{- end }}\n", nestedIndent)
			default:
				fmt.Fprintf(b, "%s{{- if index .%s 0 \"%s\" }}\n", nestedIndent, field, nestedField)
				fmt.Fprintf(b, "%s%s = {{ renderValue (index .%s 0 \"%s\") }}\n", nestedIndent, nestedField, field, nestedField)
				fmt.Fprintf(b, "%s{{- end }}\n", nestedIndent)
			}
		}

		fmt.Fprintf(b, "%s}\n", indent)
		fmt.Fprintf(b, "%s{{- end }}\n", indent)

	case *schema.Schema:
		fmt.Fprintf(b, "%s{{- if .%s }}\n", indent, field)
		fmt.Fprintf(b, "%s%s = [\n", indent, field)
		fmt.Fprintf(b, "%s  {{- range $idx, $item := .%s }}\n", indent, field)
		fmt.Fprintf(b, "%s  {{ renderValue $item }},\n", indent)
		fmt.Fprintf(b, "%s  {{- end }}\n", indent)
		fmt.Fprintf(b, "%s]\n", indent)
		fmt.Fprintf(b, "%s{{- end }}\n", indent)
	}
}

func (g *TextTemplateGenerator) generateListBlock(b *strings.Builder, parentField, field string, res *schema.Resource, indent string) {
	fmt.Fprintf(b, "%s{{- if index .%s 0 \"%s\" }}\n", indent, parentField, field)
	fmt.Fprintf(b, "%s%s {\n", indent, field)

	nestedIndent := indent + "  "

	var fields []string
	for k, v := range res.Schema {
		// Use same filtering logic as in generateNestedBlock
		if !v.Computed || v.Optional {
			fields = append(fields, k)
		}
	}
	sort.Strings(fields)

	for _, nestedField := range fields {
		nestedSchema := res.Schema[nestedField]

		if nestedSchema.Type == schema.TypeBool {
			fmt.Fprintf(b, "%s{{- if ne (index .%s 0 \"%s\" 0 \"%s\") nil }}\n",
				nestedIndent, parentField, field, nestedField)
			fmt.Fprintf(b, "%s%s = {{ index .%s 0 \"%s\" 0 \"%s\" }}\n",
				nestedIndent, nestedField, parentField, field, nestedField)
			fmt.Fprintf(b, "%s{{- end }}\n", nestedIndent)
		} else {
			fmt.Fprintf(b, "%s{{- if index .%s 0 \"%s\" 0 \"%s\" }}\n",
				nestedIndent, parentField, field, nestedField)
			fmt.Fprintf(b, "%s%s = {{ renderValue (index .%s 0 \"%s\" 0 \"%s\") }}\n",
				nestedIndent, nestedField, parentField, field, nestedField)
			fmt.Fprintf(b, "%s{{- end }}\n", nestedIndent)
		}
	}

	fmt.Fprintf(b, "%s}\n", indent)
	fmt.Fprintf(b, "%s{{- end }}\n", indent)
}

func (g *TextTemplateGenerator) generatePrimitiveList(b *strings.Builder, parentField, field string, _ *schema.Schema, indent string) {
	fmt.Fprintf(b, "%s{{- if index .%s 0 \"%s\" }}\n", indent, parentField, field)
	fmt.Fprintf(b, "%s%s = [\n", indent, field)
	fmt.Fprintf(b, "%s  {{- range $idx, $item := index .%s 0 \"%s\" }}\n", indent, parentField, field)
	fmt.Fprintf(b, "%s  {{ renderValue $item }},\n", indent)
	fmt.Fprintf(b, "%s  {{- end }}\n", indent)
	fmt.Fprintf(b, "%s]\n", indent)
	fmt.Fprintf(b, "%s{{- end }}\n", indent)
}
