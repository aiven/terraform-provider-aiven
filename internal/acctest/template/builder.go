package template

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// CompositionBuilder helps build complex compositions of templates
type CompositionBuilder struct {
	registry     *registry
	compositions []compositionEntry
}

// AddTemplate adds a new template and config to the composition
func (b *CompositionBuilder) AddTemplate(templateKey string, cfg map[string]any) *CompositionBuilder {
	b.compositions = append(b.compositions, compositionEntry{
		TemplateKey:  templateKey,
		Config:       cfg,
		ResourceType: templateKey,          // for custom templates, use template key as resource type
		ResourceKind: ResourceKindResource, // default to Resource kind for custom templates
	})
	return b
}

// AddResource helper methods to explicitly add resources or data sources
func (b *CompositionBuilder) AddResource(resourceType string, cfg map[string]any) *CompositionBuilder {
	b.compositions = append(b.compositions, compositionEntry{
		TemplateKey:  templateKey(resourceType, ResourceKindResource),
		Config:       processConfig(cfg),
		ResourceType: resourceType,
		ResourceKind: ResourceKindResource,
	})
	return b
}

func (b *CompositionBuilder) AddDataSource(resourceType string, cfg map[string]any) *CompositionBuilder {
	b.compositions = append(b.compositions, compositionEntry{
		TemplateKey:  templateKey(resourceType, ResourceKindDataSource),
		Config:       processConfig(cfg),
		ResourceType: resourceType,
		ResourceKind: ResourceKindDataSource,
	})

	return b
}

// Remove removes the template from the composition by its resource path
func (b *CompositionBuilder) Remove(resourcePath string) *CompositionBuilder {
	// Split resource path into type and name (e.g., "aiven_kafka_topic.example")
	parts := strings.Split(resourcePath, ".")
	if len(parts) != 2 {
		return b
	}
	resourceType := parts[0]

	// Filter out the composition entry that matches both resource type and name
	filtered := make([]compositionEntry, 0, len(b.compositions))
	for _, comp := range b.compositions {
		// Skip if this is the entry we want to remove
		if comp.ResourceType == resourceType {
			if name, ok := comp.Config["resource_name"].(string); ok && name == parts[1] {
				continue
			}
		}
		filtered = append(filtered, comp)
	}

	b.compositions = filtered

	return b
}

// Replace replaces the configuration of a resource in the composition
func (b *CompositionBuilder) Replace(resourcePath string, cfg map[string]any) *CompositionBuilder {
	// Split resource path into type and name
	parts := strings.Split(resourcePath, ".")
	if len(parts) != 2 {
		return b
	}
	resourceType := parts[0]
	resourceName := parts[1]

	// Process the new configuration
	processedCfg := processConfig(cfg)

	// Make sure we preserve the resource_name
	processedCfg["resource_name"] = resourceName

	// Try to find and replace the existing configuration
	found := false
	for i, comp := range b.compositions {
		if comp.ResourceType == resourceType {
			if name, ok := comp.Config["resource_name"].(string); ok && name == resourceName {
				// Preserve the template key and resource type while updating the config
				b.compositions[i].Config = processedCfg
				found = true
				break
			}
		}
	}

	// If not found, add as new
	if !found {
		b.compositions = append(b.compositions, compositionEntry{
			TemplateKey:  templateKey(resourceType, ResourceKindResource),
			Config:       processedCfg,
			ResourceType: resourceType,
			ResourceKind: ResourceKindResource,
		})
	}

	return b
}

// Clone creates a copy of the builder for branching configurations
func (b *CompositionBuilder) Clone() *CompositionBuilder {
	newCompositions := make([]compositionEntry, len(b.compositions))
	copy(newCompositions, b.compositions)

	return &CompositionBuilder{
		registry:     b.registry,
		compositions: newCompositions,
	}
}

// Factory converts a CompositionBuilder into a factory function that produces
// fresh builder instances with the same base configuration.
//
// When testing with Terraform's resource.ParallelTest, all step configurations
// are evaluated at test setup time, before any steps actually run. Without proper
// isolation, test steps using the same builder instance interfere with each other,
// as earlier steps' configuration changes persist in the builder when later steps
// are being prepared.
//
// This method solves taking a snapshot of the current builder state
// and providing a function that returns a fresh copy of the builder with the same
// base configuration.
//
// Example usage:
//
//	baseBuilder := InitializeTemplateStore(t).NewBuilder().
//	    AddResource("common_resource", {...}).
//	    AddDataSource("required_datasource", {...})
//
//	templFactory := baseBuilder.Factory()
//
//	// Each test step gets a fresh builder with the common resources already configured
//	Config: templFactory().
//	    AddResource("test_resource", {...}).
//	    MustRender(t)
func (b *CompositionBuilder) Factory() func() *CompositionBuilder {
	snapshot := b.Clone()

	return func() *CompositionBuilder {
		return snapshot.Clone()
	}
}

func (b *CompositionBuilder) Clear() *CompositionBuilder {
	b.compositions = make([]compositionEntry, 0)

	return b
}

// compositionEntry represents a combination of template and its config
type compositionEntry struct {
	TemplateKey  string
	Config       map[string]any
	ResourceType string // AddTemplate this to track the resource type
	ResourceKind ResourceKind
}

// formatTemplateError creates a user-friendly error message for template errors
func formatTemplateError(err error, templateKey string, cfg map[string]any) string {
	errStr := err.Error()

	// Handle missing resource case
	if strings.Contains(errStr, "does not exist for resource") {
		resourceType := strings.TrimPrefix(templateKey, "resource.")
		resourceType = strings.TrimPrefix(resourceType, "data.")
		return fmt.Sprintf("Resource type %q not found. Please check if the resource type is correct.", resourceType)
	}

	// Handle missing required field case
	if strings.Contains(errStr, "error calling required") {
		// Extract field name from error message
		// Example: "... at <required .project>: error calling required ..."
		field := ""
		if matches := regexp.MustCompile(`required \.([\w_]+)`).FindStringSubmatch(errStr); len(matches) > 1 {
			field = matches[1]
		}

		resourceName := "unknown"
		if name, ok := cfg["resource_name"]; ok && name != nil {
			resourceName = fmt.Sprintf("%v", name)
		}

		if field != "" {
			return fmt.Sprintf("Missing required field %q for key: %q off resource: %q", field, templateKey, resourceName)
		}
		return fmt.Sprintf("Missing required field %q for key: %q off resource: %q", field, templateKey, resourceName)
	}

	// Handle template syntax errors
	if strings.Contains(errStr, "template:") && strings.Contains(errStr, "executing") {
		return "Invalid configuration for resource. Please check all required fields are provided."
	}

	return fmt.Sprintf("Configuration error: %s", errStr)
}

func processConfig(cfg map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range cfg {
		switch val := v.(type) {
		case []map[string]any:
			// Handle list of maps
			processed := make([]map[string]any, len(val))
			for i, m := range val {
				processed[i] = processConfig(m)
			}
			result[k] = processed
		case map[string]any:
			// Handle nested maps
			result[k] = processConfig(val)
		case []any:
			// Handle mixed arrays
			processed := make([]any, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]any); ok {
					processed[i] = processConfig(m)
				} else {
					processed[i] = item
				}
			}
			result[k] = processed
		case []string:
			// Handle string arrays explicitly
			result[k] = val
		case bool:
			// Handle boolean values explicitly
			result[k] = val
		default:
			// Handle all other values
			if val != nil {
				result[k] = val
			}
		}
	}
	return result
}

// Render renders all templates in the composition and combines them
func (b *CompositionBuilder) Render(t testing.TB) (string, error) {
	t.Helper()

	if len(b.compositions) == 0 {
		return "", fmt.Errorf("no templates added to composition")
	}

	var (
		renderedParts = make([]string, 0, len(b.compositions))
		errors        []string
	)

	// render each template
	for _, comp := range b.compositions {
		rendered, err := b.registry.render(comp.TemplateKey, comp.Config)
		if err != nil {
			errorMsg := formatTemplateError(err, comp.TemplateKey, comp.Config)
			errors = append(errors, errorMsg)
			continue
		}
		renderedParts = append(renderedParts, rendered)
	}

	// If we encountered any errors, return them all
	if len(errors) > 0 {
		return "", fmt.Errorf("configuration Error(s):\n%s\n\n%s",
			strings.Join(errors, "\n"),
			b.debugString())
	}

	// Combine all rendered parts
	combined := strings.Join(renderedParts, "\n\n")

	// Do a simple HCL validation of the final result
	if err := validateHCL(combined); err != nil {
		return "", fmt.Errorf("invalid HCL generated:\n%s\n\nError: %w", combined, err)
	}

	return combined, nil
}

// MustRender is like Render but fails the test on error with detailed message
func (b *CompositionBuilder) MustRender(t testing.TB) string {
	t.Helper()

	result, err := b.Render(t)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// Simple HCL validation function
func validateHCL(content string) error {
	_, diags := hclwrite.ParseConfig([]byte(content), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("%w", diags)
	}

	return nil
}

// resourceDescription returns a readable description of the resource
func resourceDescription(comp compositionEntry) string {
	resourceName, ok := comp.Config["resource_name"]
	if !ok || resourceName == nil {
		return comp.ResourceType
	}

	return fmt.Sprintf("%s (name: %v)", comp.ResourceType, resourceName)
}

// debugString returns a concise debug representation of the composition
func (b *CompositionBuilder) debugString() string {
	var buf strings.Builder
	buf.WriteString("Current Configuration:\n")

	// Group by kind
	var (
		dataResources []string
		resources     []string
		others        []string
	)

	for _, comp := range b.compositions {
		desc := resourceDescription(comp)
		switch comp.ResourceKind {
		case ResourceKindDataSource:
			dataResources = append(dataResources, desc)
		case ResourceKindResource:
			resources = append(resources, desc)
		default:
			others = append(others, desc)
		}
	}

	if len(dataResources) > 0 {
		buf.WriteString("\nData Sources:\n")
		for _, r := range dataResources {
			buf.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	if len(resources) > 0 {
		buf.WriteString("\nResources:\n")
		for _, r := range resources {
			buf.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	if len(others) > 0 {
		buf.WriteString("\nOther:\n")
		for _, r := range others {
			buf.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	return buf.String()
}
