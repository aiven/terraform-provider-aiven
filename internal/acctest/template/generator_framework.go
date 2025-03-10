package template

import (
	"fmt"
	"strings"
)

// FrameworkTemplateGenerator implements TemplateGenerator for Terraform Plugin Framework resources
type FrameworkTemplateGenerator struct {
	extractor *FrameworkFieldExtractor
	renderer  *CommonTemplateRenderer
}

// NewFrameworkTemplateGenerator creates a template generator for Framework resources
func NewFrameworkTemplateGenerator() *FrameworkTemplateGenerator {
	return &FrameworkTemplateGenerator{
		extractor: NewFrameworkFieldExtractor(),
		renderer:  &CommonTemplateRenderer{},
	}
}

// GenerateTemplate generates a Go template string that resembles Terraform HCL configuration
// for a Framework resource or data source.
func (g *FrameworkTemplateGenerator) GenerateTemplate(schema interface{}, resourceType string, kind ResourceKind) string {
	var b strings.Builder

	// Header differs based on kind
	_, _ = fmt.Fprintf(&b, "%s %q %q {\n", kind, resourceType, "{{ required .resource_name }}")

	fields, err := g.extractor.ExtractFields(schema)
	if err != nil {
		// If schema type is not recognized, return a basic template with an error comment
		b.WriteString(fmt.Sprintf("  # Error extracting fields: %v\n", err))
		b.WriteString("}")
		return b.String()
	}

	var (
		hasTimeouts   bool
		hasDependsOn  bool
		timeoutsField TemplateField
		regularFields []TemplateField
	)

	for _, field := range fields {
		switch field.Name {
		case "timeouts":
			hasTimeouts = true
			timeoutsField = field
		case "depends_on":
			hasDependsOn = true
		default:
			regularFields = append(regularFields, field)
		}
	}

	for _, field := range regularFields {
		g.renderer.RenderField(&b, field, 1, TemplatePath{})
	}

	if hasTimeouts {
		timeoutsConfig := g.extractTimeoutsConfig(timeoutsField)
		g.renderer.RenderTimeouts(&b, 1, timeoutsConfig)
	}

	if hasDependsOn {
		g.renderer.RenderDependsOn(&b, 1)
	}

	b.WriteString("}")

	return b.String()
}

// extractTimeoutsConfig analyzes the timeouts field to determine which timeouts are configured
func (g *FrameworkTemplateGenerator) extractTimeoutsConfig(timeoutsField TemplateField) TimeoutsConfig {
	config := TimeoutsConfig{
		Create: false,
		Read:   false,
		Update: false,
		Delete: false,
	}

	for _, nestedField := range timeoutsField.NestedFields {
		fieldName := strings.ToLower(nestedField.Name)
		switch fieldName {
		case "create":
			config.Create = true
		case "read":
			config.Read = true
		case "update":
			config.Update = true
		case "delete":
			config.Delete = true
		}
	}

	return config
}
