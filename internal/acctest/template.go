package acctest

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"testing"
)

// ResourceConfig is the interface that all resource configs must implement
type resourceConfig interface {
	// ToMap converts the config to a map for template rendering
	ToMap() map[string]any
}

// Template represents a single Terraform configuration template
type Template struct {
	Name     string
	Template string
}

// TemplateRegistry holds templates for a specific resource type
type TemplateRegistry struct {
	resourceName string
	templates    map[string]*template.Template
	funcMap      template.FuncMap
}

// NewTemplateRegistry creates a new template registry for a resource
func NewTemplateRegistry(resourceName string) *TemplateRegistry {
	return &TemplateRegistry{
		resourceName: resourceName,
		templates:    make(map[string]*template.Template),
		funcMap:      make(template.FuncMap),
	}
}

// AddTemplate adds a new template to the registry
func (r *TemplateRegistry) AddTemplate(t testing.TB, name, templateStr string) error {
	t.Helper()

	tmpl := template.New(name)
	if len(r.funcMap) > 0 {
		tmpl = tmpl.Funcs(r.funcMap)
	}

	parsed, err := tmpl.Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	r.templates[name] = parsed

	return nil
}

// MustAddTemplate is like AddTemplate but panics on error
func (r *TemplateRegistry) MustAddTemplate(t testing.TB, name, templateStr string) {
	t.Helper()

	if err := r.AddTemplate(t, name, templateStr); err != nil {
		t.Fatal(err)
	}
}

// Render renders a template with the given config
func (r *TemplateRegistry) Render(t testing.TB, templateKey string, cfg map[string]any) (string, error) {
	t.Helper()

	tmpl, exists := r.templates[templateKey]
	if !exists {
		availableTemplates := r.getAvailableTemplates()

		return "", fmt.Errorf("template %q does not exist for resource %s. Available templates: %v",
			templateKey,
			r.resourceName,
			availableTemplates,
		)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// MustRender is like Render but fails the test on error
func (r *TemplateRegistry) MustRender(t testing.TB, templateKey string, cfg map[string]any) string {
	t.Helper()

	result, err := r.Render(t, templateKey, cfg)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// AddFunction adds a custom function to the template registry
func (r *TemplateRegistry) AddFunction(name string, fn interface{}) {
	if r.funcMap == nil {
		r.funcMap = make(template.FuncMap)
	}
	r.funcMap[name] = fn
}

// HasTemplate checks if a template exists in the registry
func (r *TemplateRegistry) HasTemplate(key string) bool {
	_, exists := r.templates[key]
	return exists
}

// RemoveTemplate removes a template from the registry
func (r *TemplateRegistry) RemoveTemplate(key string) {
	delete(r.templates, key)
}

// getAvailableTemplates returns a sorted list of available template keys
func (r *TemplateRegistry) getAvailableTemplates() []string {
	templates := make([]string, 0, len(r.templates))
	for k := range r.templates {
		templates = append(templates, k)
	}
	sort.Strings(templates)

	return templates
}

// compositionEntry represents a combination of template and its config
type compositionEntry struct {
	TemplateKey string
	Config      map[string]any
}

// CompositionBuilder helps build complex compositions of templates
type CompositionBuilder struct {
	registry     *TemplateRegistry
	compositions []compositionEntry
}

// NewCompositionBuilder creates a new composition builder
func (r *TemplateRegistry) NewCompositionBuilder() *CompositionBuilder {
	return &CompositionBuilder{
		registry:     r,
		compositions: make([]compositionEntry, 0),
	}
}

// Add adds a new template and config to the composition
func (b *CompositionBuilder) Add(templateKey string, cfg map[string]any) *CompositionBuilder {
	b.compositions = append(b.compositions, compositionEntry{
		TemplateKey: templateKey,
		Config:      cfg,
	})
	return b
}

// AddWithConfig adds a new template and config to the composition using a resourceConfig
func (b *CompositionBuilder) AddWithConfig(templateKey string, cfg resourceConfig) *CompositionBuilder {
	b.compositions = append(b.compositions, compositionEntry{
		TemplateKey: templateKey,
		Config:      cfg.ToMap(),
	})
	return b
}

// AddIf conditional method to CompositionBuilder
func (b *CompositionBuilder) AddIf(condition bool, templateKey string, cfg map[string]any) *CompositionBuilder {
	if condition {
		return b.Add(templateKey, cfg)
	}

	return b
}

func (b *CompositionBuilder) Remove(templateKey string) *CompositionBuilder {
	var newCompositions []compositionEntry
	for _, comp := range b.compositions {
		if comp.TemplateKey != templateKey {
			newCompositions = append(newCompositions, comp)
		}
	}
	b.compositions = newCompositions

	return b
}

// Render renders all templates in the composition and combines them
func (b *CompositionBuilder) Render(t testing.TB) (string, error) {
	t.Helper()

	var renderedParts = make([]string, 0, len(b.compositions))

	// Render each template
	for _, comp := range b.compositions {
		rendered, err := b.registry.Render(t, comp.TemplateKey, comp.Config)
		if err != nil {
			return "", fmt.Errorf("failed to render template %s: %w", comp.TemplateKey, err)
		}
		renderedParts = append(renderedParts, rendered)
	}

	// Combine all rendered parts
	combined := strings.Join(renderedParts, "\n\n")

	//TODO: add HCL validation?

	return combined, nil
}

// MustRender is like Render but fails the test on error
func (b *CompositionBuilder) MustRender(t testing.TB) string {
	t.Helper()

	result, err := b.Render(t)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
