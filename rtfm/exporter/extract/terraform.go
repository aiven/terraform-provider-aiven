package extract

import (
	"context"
	"fmt"
	"os"

	"github.com/aiven/terraform-provider-aiven/rtfm/exporter"
)

// TerraformExtractor export existing resources using Terraform
type TerraformExtractor struct {
	workDir       string
	cfg           *Config
	commander     Commander
	importBuilder Importer
}

func NewTerraformExecutor(cfg *Config) (*TerraformExtractor, error) {
	// check if specified terraform binary exists
	info, err := os.Stat(cfg.TerraformBinaryPath)
	if err != nil {
		return nil, fmt.Errorf("specified terraform binary not found at %s: %w", cfg.TerraformBinaryPath, err)
	}

	// check if it's executable
	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("terraform binary at %s is not executable", cfg.TerraformBinaryPath)
	}

	// create a temporary working directory //TODO: use dev/shm ???
	workDir, err := os.MkdirTemp("", "tf-workdir-")
	if err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	return &TerraformExtractor{
		cfg:           cfg,
		workDir:       workDir,
		commander:     NewTerraformCommander(cfg.TerraformBinaryPath),
		importBuilder: &TerraformImporter{},
	}, nil
}

func (e *TerraformExtractor) Export(ctx context.Context, ID exporter.ResourceInput) (exporter.ExportResult, error) {
	var result exporter.ExportResult

	// create a new project for each export
	project, err := NewProject("terraform-project-", e.commander, e.importBuilder)
	if err != nil {
		return result, fmt.Errorf("failed to create Terraform project: %w", err)
	}
	defer project.Cleanup()

	// initialize terraform
	if err = project.Init(ctx, nil); err != nil {
		return result, fmt.Errorf("terraform init failed: %w", err)
	}

	if err = project.Import(ctx, ID, nil); err != nil {
		return result, fmt.Errorf("terraform import failed: %w", err)
	}

	hclContent, err := os.ReadFile(project.exportPath(ID))
	if err != nil {
		return result, fmt.Errorf("failed to read exported HCL file at %s: %w", project.exportPath(ID), err)
	}

	result.HCL = string(hclContent)

	return result, nil
}
