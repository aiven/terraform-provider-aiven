package template

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"testing"
)

// registry holds templates for a specific resource type
type registry struct {
	t         testing.TB
	templates map[string]*template.Template
	funcs     *templateFunctions
}

// newTemplateRegistry creates a new template registry for a resource
func newTemplateRegistry(t testing.TB) *registry {
	return &registry{
		t:         t,
		templates: make(map[string]*template.Template),
		funcs:     newTemplateFunctions(),
	}
}

// addTemplate adds a new template to the registry
func (r *registry) addTemplate(name, templateStr string) error {
	r.t.Helper()

	tmpl := template.New(name).Funcs(r.funcs.getFuncMap())

	parsed, err := tmpl.Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	r.templates[name] = parsed

	return nil
}

// mustAddTemplate is like addTemplate but panics on error
func (r *registry) mustAddTemplate(name, templateStr string) {
	r.t.Helper()

	if err := r.addTemplate(name, templateStr); err != nil {
		r.t.Fatalf("failed to add template %q: %s", templateStr, err)
	}
}

// render renders a template with the given config
func (r *registry) render(templateKey string, cfg map[string]any) (string, error) {
	r.t.Helper()

	tmpl, exists := r.templates[templateKey]
	if !exists {
		availableTemplates := r.getAvailableTemplates()
		return "", fmt.Errorf("template %q does not exist for resource. Available templates: %v",
			templateKey,
			availableTemplates,
		)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// addFunction adds a custom function to the template registry
func (r *registry) addFunction(name string, fn any) {
	r.funcs.register(name, fn)
}

// getAvailableTemplates returns a sorted list of available template keys
func (r *registry) getAvailableTemplates() []string {
	templates := make([]string, 0, len(r.templates))
	for k := range r.templates {
		templates = append(templates, k)
	}
	sort.Strings(templates)

	return templates
}
