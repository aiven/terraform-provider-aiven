package extract

import (
	"context"
	"fmt"

	"github.com/aiven/terraform-provider-aiven/rtfm/exporter"
)

type Importer interface {
	// BuildCommand builds the Terraform import command for the given resource input
	BuildCommand(ctx context.Context, id exporter.ResourceInput) (string, error)
}

type TerraformImporter struct {
}

func (i *TerraformImporter) BuildCommand(_ context.Context, id exporter.ResourceInput) (string, error) {
	switch id.Type {
	case exporter.ResourceTypeKafka, exporter.ResourceTypePG:
		resourceAddress := fmt.Sprintf("%s.%s", id.Type, "example")
		resourceID := fmt.Sprintf("%s/%s", id.Project, id.ServiceName)

		return fmt.Sprintf("%s %s", resourceAddress, resourceID), nil
	default:
		return "", fmt.Errorf("unsupported resource type for import command: %s", id.Type)
	}
}
