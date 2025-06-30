package exporter

import (
	"context"
)

// SupportedResourceType defines the type of Aiven resource
type SupportedResourceType string

const (
	ResourceTypeKafka SupportedResourceType = "aiven_kafka"
	ResourceTypePG    SupportedResourceType = "aiven_pg"
)

type ResourceInput struct {
	Type        SupportedResourceType
	Project     string
	ServiceName string
	Token       string

	Attributes map[string]any
}

type ExportResult struct {
	HCL           string
	ImportCommand string
}

type ResourceExporter interface {
	Export(ctx context.Context, ID ResourceInput) (ExportResult, error)
}

func IsSupportedResourceType(resourceType string) bool {
	switch SupportedResourceType(resourceType) {
	case ResourceTypeKafka, ResourceTypePG:
		return true
	default:
		return false
	}
}
