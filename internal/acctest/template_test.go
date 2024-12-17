package acctest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompositionBuilder_SingleResource(t *testing.T) {
	registry := NewTemplateRegistry("test")

	template := `resource "aiven_project" "example_project" {
  project = "{{ .project_name }}"
}`
	registry.MustAddTemplate(t, "project", template)

	builder := registry.NewCompositionBuilder()
	builder.Add("project", map[string]any{
		"project_name": "test-project",
	})

	expected := `resource "aiven_project" "example_project" {
  project = "test-project"
}`

	result := builder.MustRender(t)
	assert.Equal(t, normalizeHCL(expected), normalizeHCL(result))
}

func TestCompositionBuilder_TwoIndependentResources(t *testing.T) {
	registry := NewTemplateRegistry("test")

	registry.MustAddTemplate(t, "org_unit", `resource "aiven_organizational_unit" "example_unit" {
  name = "{{ .name }}"
}`)
	registry.MustAddTemplate(t, "billing_group", `resource "aiven_billing_group" "example_billing_group" {
  name             = "{{ .name }}"
  billing_currency = "{{ .currency }}"
}`)

	builder := registry.NewCompositionBuilder()
	builder.Add("org_unit", map[string]any{
		"name": "Example Unit",
	})
	builder.Add("billing_group", map[string]any{
		"name":     "Example Billing",
		"currency": "USD",
	})

	expected := `resource "aiven_organizational_unit" "example_unit" {
  name = "Example Unit"
}
resource "aiven_billing_group" "example_billing_group" {
  name             = "Example Billing"
  billing_currency = "USD"
}`

	result := builder.MustRender(t)
	assert.Equal(t, normalizeHCL(expected), normalizeHCL(result))
}

func TestCompositionBuilder_DependentResources(t *testing.T) {
	registry := NewTemplateRegistry("test")

	registry.MustAddTemplate(t, "billing_group", `resource "aiven_billing_group" "example_billing_group" {
  name             = "{{ .name }}"
  billing_currency = "USD"
}`)
	registry.MustAddTemplate(t, "project", `resource "aiven_project" "example_project" {
  project       = "{{ .name }}"
  billing_group = aiven_billing_group.example_billing_group.id
}`)

	builder := registry.NewCompositionBuilder()
	builder.Add("billing_group", map[string]any{
		"name": "example-billing",
	})
	builder.Add("project", map[string]any{
		"name": "example-project",
	})

	expected := `resource "aiven_billing_group" "example_billing_group" {
  name             = "example-billing"
  billing_currency = "USD"
}
resource "aiven_project" "example_project" {
  project       = "example-project"
  billing_group = aiven_billing_group.example_billing_group.id
}`

	result := builder.MustRender(t)
	assert.Equal(t, normalizeHCL(expected), normalizeHCL(result))
}

func TestCompositionBuilder_ConditionalResource(t *testing.T) {
	registry := NewTemplateRegistry("test")

	registry.MustAddTemplate(t, "project", `resource "aiven_project" "example_project" {
  project = "{{ .name }}"
}`)
	registry.MustAddTemplate(t, "redis", `resource "aiven_redis" "redis1" {
  project      = aiven_project.example_project.project
  service_name = "{{ .name }}"
  plan         = "{{ .plan }}"
}`)

	tests := []struct {
		name           string
		includeRedis   bool
		expectedOutput string
	}{
		{
			name:         "with_redis",
			includeRedis: true,
			expectedOutput: `resource "aiven_project" "example_project" {
  project = "test-project"
}
resource "aiven_redis" "redis1" {
  project      = aiven_project.example_project.project
  service_name = "test-redis"
  plan         = "business-4"
}`,
		},
		{
			name:         "without_redis",
			includeRedis: false,
			expectedOutput: `resource "aiven_project" "example_project" {
  project = "test-project"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := registry.NewCompositionBuilder()

			builder.Add("project", map[string]any{
				"name": "test-project",
			})

			builder.AddIf(tt.includeRedis, "redis", map[string]any{
				"name": "test-redis",
				"plan": "business-4",
			})

			result := builder.MustRender(t)
			assert.Equal(t, normalizeHCL(tt.expectedOutput), normalizeHCL(result))
		})
	}
}

func TestCompositionBuilder_DataSourceAndResource(t *testing.T) {
	registry := NewTemplateRegistry("test")

	registry.MustAddTemplate(t, "org_data", `data "aiven_organization" "main" {
  name = "{{ .name }}"
}`)
	registry.MustAddTemplate(t, "project", `resource "aiven_project" "example_project" {
  project         = "{{ .name }}"
  organization_id = data.aiven_organization.main.id
}`)

	builder := registry.NewCompositionBuilder()
	builder.Add("org_data", map[string]any{
		"name": "example-org",
	})
	builder.Add("project", map[string]any{
		"name": "example-project",
	})

	expected := `data "aiven_organization" "main" {
  name = "example-org"
}
resource "aiven_project" "example_project" {
  project         = "example-project"
  organization_id = data.aiven_organization.main.id
}`

	result := builder.MustRender(t)
	assert.Equal(t, normalizeHCL(expected), normalizeHCL(result))
}

func TestTemplateNilAndMissingValues(t *testing.T) {
	registry := NewTemplateRegistry("test")

	templateStr := `resource "aiven_service" "example" {
  project      = "{{ .project }}"
  service_name = "{{ .service_name }}"
  {{- if .maintenance_window_dow }}
  maintenance_window_dow = "{{ .maintenance_window_dow }}"
  {{- end }}
  {{- if .maintenance_window_time }}
  maintenance_window_time = "{{ .maintenance_window_time }}"
  {{- end }}
}`

	registry.MustAddTemplate(t, "service", templateStr)

	tests := []struct {
		name     string
		config   map[string]any
		expected string
	}{
		{
			name: "explicit_nil_value",
			config: map[string]any{
				"project":                 "test-project",
				"service_name":            "test-service",
				"maintenance_window_dow":  nil,
				"maintenance_window_time": "10:00:00",
			},
			expected: `resource "aiven_service" "example" {
  project                 = "test-project"
  service_name            = "test-service"
  maintenance_window_time = "10:00:00"
}`,
		},
		{
			name: "missing_key",
			config: map[string]any{
				"project":      "test-project",
				"service_name": "test-service",
				// maintenance_window_dow is not in config at all
				"maintenance_window_time": "10:00:00",
			},
			expected: `resource "aiven_service" "example" {
  project                 = "test-project"
  service_name            = "test-service"
  maintenance_window_time = "10:00:00"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.MustRender(t, "service", tt.config)
			assert.Equal(t, normalizeHCL(tt.expected), normalizeHCL(result))
		})
	}
}

func TestCompositionBuilder(t *testing.T) {
	tests := []struct {
		name         string
		templates    map[string]string
		compositions []struct {
			templateKey string
			config      map[string]any
		}
		expectedOutput string
		expectError    bool
	}{
		{
			name: "kafka_with_user",
			templates: map[string]string{
				"kafka": `resource "aiven_kafka" "example_kafka" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "{{ .cloud_name }}"
  plan                    = "{{ .plan }}"
  service_name            = "{{ .service_name }}"
  maintenance_window_dow  = "{{ .maintenance_window_dow }}"
  maintenance_window_time = "{{ .maintenance_window_time }}"
  kafka_user_config {
    kafka_rest      = "{{ .kafka_rest }}"
    kafka_connect   = "{{ .kafka_connect }}"
    schema_registry = "{{ .schema_registry }}"
    kafka_version   = "{{ .kafka_version }}"
    kafka {
      group_max_session_timeout_ms = "{{ .group_max_session_timeout_ms }}"
      log_retention_bytes          = "{{ .log_retention_bytes }}"
    }
    public_access {
      kafka_rest    = "{{ .kafka_rest_public }}"
      kafka_connect = "{{ .kafka_connect_public }}"
    }
  }
}`,
				"kafka_user": `resource "aiven_kafka_user" "example_service_user" {
  service_name = aiven_kafka.example_kafka.service_name
  project      = data.aiven_project.example_project.project
  username     = "{{ .username }}"
  password     = "{{ .password }}"
}`,
				"project_data": `data "aiven_project" "example_project" {
  project = "{{ .project }}"
}`,
			},
			compositions: []struct {
				templateKey string
				config      map[string]any
			}{
				{
					templateKey: "project_data",
					config: map[string]any{
						"project": "example-project",
					},
				},
				{
					templateKey: "kafka",
					config: map[string]any{
						"cloud_name":                   "google-europe-west1",
						"plan":                         "business-4",
						"service_name":                 "example-kafka",
						"maintenance_window_dow":       "monday",
						"maintenance_window_time":      "10:00:00",
						"kafka_rest":                   true,
						"kafka_connect":                true,
						"schema_registry":              true,
						"kafka_version":                "3.5",
						"group_max_session_timeout_ms": 70000,
						"log_retention_bytes":          1000000000,
						"kafka_rest_public":            true,
						"kafka_connect_public":         true,
					},
				},
				{
					templateKey: "kafka_user",
					config: map[string]any{
						"username": "example-kafka-user",
						"password": "dummy-password",
					},
				},
			},
			expectedOutput: `data "aiven_project" "example_project" {
  project = "example-project"
}
resource "aiven_kafka" "example_kafka" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "example-kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka_rest      = "true"
    kafka_connect   = "true"
    schema_registry = "true"
    kafka_version   = "3.5"
    kafka {
      group_max_session_timeout_ms = "70000"
      log_retention_bytes          = "1000000000"
    }
    public_access {
      kafka_rest    = "true"
      kafka_connect = "true"
    }
  }
}
resource "aiven_kafka_user" "example_service_user" {
  service_name = aiven_kafka.example_kafka.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-kafka-user"
  password     = "dummy-password"
}`,
		},
		{
			name: "conditional_kafka_config",
			templates: map[string]string{
				"kafka_base": `resource "aiven_kafka" "kafka" {
  project      = "{{ .project }}"
  service_name = "{{ .service_name }}"
  cloud_name   = "{{ .cloud_name }}"
  plan         = "{{ .plan }}"
}`,
				"kafka_config": `  kafka_user_config {
   kafka_rest      = {{ .kafka_rest }}
   schema_registry = {{ .schema_registry }}
 }`,
			},
			compositions: []struct {
				templateKey string
				config      map[string]any
			}{
				{
					templateKey: "kafka_base",
					config: map[string]any{
						"project":      "test-project",
						"service_name": "test-kafka",
						"cloud_name":   "google-europe-west1",
						"plan":         "business-4",
					},
				},
			},
			expectedOutput: `resource "aiven_kafka" "kafka" {
  project      = "test-project"
  service_name = "test-kafka"
  cloud_name   = "google-europe-west1"
  plan         = "business-4"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTemplateRegistry("test")

			// Add all templates
			for key, tmpl := range tt.templates {
				err := registry.AddTemplate(t, key, tmpl)
				assert.NoError(t, err)
			}

			// Create composition
			builder := registry.NewCompositionBuilder()
			for _, comp := range tt.compositions {
				builder.Add(comp.templateKey, comp.config)
			}

			// Render and verify
			result, err := builder.Render(t)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t,
				normalizeHCL(tt.expectedOutput),
				normalizeHCL(result),
				"Rendered template should match expected output",
			)
		})
	}
}

// normalizeHCL function remains the same
func normalizeHCL(s string) string {
	// Split into lines for processing
	lines := strings.Split(s, "\n")
	var normalized = make([]string, 0, len(lines))

	for _, line := range lines {
		// Trim spaces from both ends
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle lines with just closing braces
		if line == "}" {
			normalized = append(normalized, line)
			continue
		}

		// For lines with content, normalize internal spacing
		if strings.Contains(line, "=") {
			// Split by = and trim spaces
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Reconstruct with consistent spacing
				line = fmt.Sprintf("  %s = %s", key, value)
			}
		} else if strings.HasPrefix(line, "resource") || strings.HasPrefix(line, "data") {
			// Handle resource and data block declarations
			line = strings.TrimSpace(line)
		} else if !strings.HasPrefix(line, "}") {
			// Add consistent indentation for other non-empty lines
			line = "  " + strings.TrimSpace(line)
		}

		normalized = append(normalized, line)
	}

	// Join lines with newlines
	result := strings.Join(normalized, "\n")

	// Normalize line endings
	result = strings.ReplaceAll(result, "\r\n", "\n")

	return result
}
