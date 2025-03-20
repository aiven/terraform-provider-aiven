package provision

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aiven/terraform-provider-aiven/rtfm/server"
)

// Request represents the incoming request for provisioning
type Request struct {
	Project      string         `json:"project"`
	ResourceType string         `json:"resource_type"`
	ResourceName string         `json:"resource_name"` // Name for the resource in Terraform config
	Attributes   map[string]any `json:"attributes"`    // Resource attributes
}

// Response represents the response to a provision request
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	HCL     string `json:"hcl,omitempty"`
}

// Handler handles provisioning requests. Temporary implementation just to demonstrate the main idea.
type Handler struct {
	templatesDir  string
	templateCache map[string]*template.Template
	templateFuncs template.FuncMap
}

// NewHandler creates a new handler for provisioning resources
func NewHandler(cfg *Config) *Handler {
	// template functions for rendering values
	funcMap := template.FuncMap{
		"renderValue": func(v any) string {
			switch val := v.(type) {
			case string:
				return fmt.Sprintf("%q", val)
			case int, int64, float64:
				return fmt.Sprintf("%v", val)
			case bool:
				return fmt.Sprintf("%v", val)
			case []any:
				var items []string
				for _, item := range val {
					items = append(items, renderValueAsString(item))
				}
				return fmt.Sprintf("[%s]", strings.Join(items, ", "))
			case map[string]any:
				var pairs []string
				for k, v := range val {
					pairs = append(pairs, fmt.Sprintf("%q = %s", k, renderValueAsString(v)))
				}
				return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
			default:
				if val == nil {
					return "null"
				}
				return fmt.Sprintf("%v", val)
			}
		},
		"required": func(v any) any {
			if v == nil {
				return fmt.Errorf("required field is missing")
			}
			return v
		},
	}

	return &Handler{
		templatesDir:  cfg.TemplatesDir,
		templateCache: make(map[string]*template.Template),
		templateFuncs: funcMap,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var frontendReq map[string]any

	// parse the request body
	if err := json.NewDecoder(r.Body).Decode(&frontendReq); err != nil {
		server.WriteJSONResponse(w, http.StatusBadRequest, server.ErrorResponse{
			Status:  "error",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	projectName := r.PathValue("project")
	if projectName == "" {
		server.WriteJSONResponse(w, http.StatusBadRequest, server.ErrorResponse{
			Status:  "error",
			Message: "Empty project name",
		})

		return
	}

	templateData, resourceType := MapServiceRequest(frontendReq, projectName)

	hcl, err := h.renderTemplate(resourceType, templateData)
	if err != nil {
		server.WriteJSONResponse(w, http.StatusInternalServerError, server.ErrorResponse{
			Status:  "error",
			Message: "Failed to render template: " + err.Error(),
		})
		return
	}

	hcl = html.UnescapeString(hcl)

	// Return the response
	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("Resource %s with name %s provisioned", resourceType, frontendReq["service_name"]),
		HCL:     hcl,
	}

	server.WriteJSONResponse(w, http.StatusOK, response)
}

// renderTemplate renders a template with the given data
func (h *Handler) renderTemplate(resourceType string, data map[string]any) (string, error) {
	tmpl, ok := h.templateCache[resourceType]
	if !ok {
		templatePath := filepath.Join(h.templatesDir, resourceType+".tpl")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			templatePath = filepath.Join(h.templatesDir, strings.ReplaceAll(resourceType, ".", "_")+".tpl")
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				return "", fmt.Errorf("template for resource type %s not found", resourceType)
			}
		}

		// read template file
		templateContent, err := os.ReadFile(templatePath)
		if err != nil {
			return "", fmt.Errorf("failed to read template file: %w", err)
		}

		// parse template
		tmpl, err = template.New(resourceType).Funcs(h.templateFuncs).Parse(string(templateContent))
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %w", err)
		}

		// cache the parsed template
		h.templateCache[resourceType] = tmpl
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// renderValueAsString helper function to render a value as a string
func renderValueAsString(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case int, int64, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	default:
		if val == nil {
			return "null"
		}
		return fmt.Sprintf("%v", val)
	}
}
