package extract

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aiven/terraform-provider-aiven/rtfm/exporter"
	"github.com/aiven/terraform-provider-aiven/rtfm/server"
)

// Request represents the incoming request for terraform import
type Request struct {
	Project      string `json:"project"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

// Response represents the response to an import request
type Response struct {
	Status        string `json:"status"`
	Message       string `json:"message"`
	HCL           string `json:"hcl,omitempty"`
	ImportCommand string `json:"import_command,omitempty"`
}

type Handler struct {
	exporter exporter.ResourceExporter
}

func NewHandler(exp exporter.ResourceExporter) *Handler {
	return &Handler{
		exporter: exp,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse := server.ErrorResponse{
			Status:  "error",
			Message: "Invalid request format: " + err.Error(),
		}
		server.WriteJSONResponse(w, http.StatusBadRequest, errorResponse)
		return
	}

	token, err := tokenFromRequest(r)
	if err != nil {
		errorResponse := server.ErrorResponse{
			Status:  "error",
			Message: "Failed to get token: " + err.Error(),
		}
		server.WriteJSONResponse(w, http.StatusForbidden, errorResponse)
		return
	}

	// validate
	if req.ResourceType == "" || req.ResourceID == "" || req.Project == "" {
		errorResponse := server.ErrorResponse{
			Status:  "error",
			Message: "Missing required fields: resource_type, resource_id, and resource_name are required",
		}
		server.WriteJSONResponse(w, http.StatusBadRequest, errorResponse)
		return
	}

	//check if the resource type is supported
	if !exporter.IsSupportedResourceType(req.ResourceType) {
		errorResponse := server.ErrorResponse{
			Status:  "error",
			Message: "Resource type is not supported",
		}
		server.WriteJSONResponse(w, http.StatusBadRequest, errorResponse)
		return
	}

	res, err := h.exporter.Export(r.Context(), exporter.ResourceInput{
		Type:        exporter.SupportedResourceType(req.ResourceType),
		Project:     req.Project,
		ServiceName: req.ResourceID,
		Token:       token,
	})

	if err != nil {
		errorResponse := server.ErrorResponse{
			Status:  "error",
			Message: "Failed to export resource: " + err.Error(),
		}
		server.WriteJSONResponse(w, http.StatusInternalServerError, errorResponse)
		return
	}

	response := Response{
		Status:        "success",
		HCL:           res.HCL,
		ImportCommand: res.ImportCommand,
	}

	server.WriteJSONResponse(w, http.StatusOK, response)
}

func tokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" {
		return "", errors.New("authorization token is empty")
	}

	return token, nil
}
