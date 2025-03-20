package extract

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aiven/terraform-provider-aiven/rtfm/exporter"
)

// Project encapsulates a Terraform project context
type Project struct {
	workDir       string
	commander     Commander
	importBuilder Importer
}

// NewProject creates a new Terraform project context using the system temp directory
func NewProject(prefix string, commander Commander, importBuilder Importer) (*Project, error) {
	tempDir, err := os.MkdirTemp("", prefix) //todo: change this
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// create the terraform subdirectory for initialization
	if err = os.MkdirAll(filepath.Join(tempDir, ".terraform"), 0755); err != nil {
		_ = os.RemoveAll(tempDir)

		return nil, fmt.Errorf("failed to create .terraform directory: %w", err)
	}

	return &Project{
		workDir:       tempDir,
		commander:     commander,
		importBuilder: importBuilder,
	}, nil
}

// Init initializes the Terraform project
func (p *Project) Init(_ context.Context, opts *CommandOptions) error {
	if opts == nil {
		opts = &CommandOptions{}
	}

	// Always use the project's work directory
	opts.WorkDir = p.workDir

	if err := p.writeAivenProvider(); err != nil {
		return fmt.Errorf("failed to write Aiven provider configuration: %w", err)
	}

	// arguments to make init faster
	// -input=false: don't ask for input
	// -lock=false: don't lock the state file
	// -upgrade=false: don't check for newer versions of providers
	initArgs := []string{
		"-input=false",
		"-lock=false",
		"-upgrade=false",
	}

	return p.commander.Execute("init", initArgs, opts)
}

// Import imports a resource into Terraform state
func (p *Project) Import(ctx context.Context, ID exporter.ResourceInput, opts *CommandOptions) error {
	if opts == nil {
		opts = &CommandOptions{}
	}

	// use the project's work directory
	opts.WorkDir = p.workDir

	// Ensure we have a placeholder configuration for this resource
	if err := p.writePlaceholderResource(ID.Type, "example"); err != nil {
		return fmt.Errorf("failed to write placeholder configuration: %w", err)
	}

	args, err := p.importBuilder.BuildCommand(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to build import command: %w", err)

	}

	env := append(
		os.Environ(),
		fmt.Sprintf("AIVEN_TOKEN=%s", ID.Token),
		fmt.Sprintf("AIVEN_PLUGIN_EXPORT_FILE_PATH=%s", p.exportPath(ID)),
	)
	opts.Env = env

	importArgs := strings.Fields(args)

	return p.commander.Execute("import", importArgs, opts)
}

// WriteConfig writes a Terraform configuration file to the project directory
func (p *Project) WriteConfig(filename string, content []byte) error {
	path := filepath.Join(p.workDir, filename)
	return os.WriteFile(path, content, 0644)
}

// Cleanup removes the project directory
func (p *Project) Cleanup() error {
	return os.RemoveAll(p.workDir)
}

func (p *Project) writeAivenProvider() error {
	// create the basic provider configuration
	providerConfig := `terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
    }
  }
}

provider "aiven" {
  # Token will be read from AIVEN_TOKEN environment variable
}
`
	return p.WriteConfig("main.tf", []byte(providerConfig))
}

// writePlaceholderResource creates a placeholder resource configuration
func (p *Project) writePlaceholderResource(resourceType exporter.SupportedResourceType, resourceName string) error {
	placeholder := fmt.Sprintf(`
resource "%s" "%s" {
  # This is a placeholder configuration for import
  # Actual values will be filled in by Terraform after import
}
`, resourceType, resourceName)

	mainTfPath := filepath.Join(p.workDir, "main.tf")
	existingContent, err := os.ReadFile(mainTfPath)
	if err != nil {
		return fmt.Errorf("failed to read main.tf: %w", err)
	}

	// append the placeholder to the existing content
	newContent := append(existingContent, []byte(placeholder)...)

	// Write the updated content back to main.tf
	return os.WriteFile(mainTfPath, newContent, 0644)
}

func (p *Project) exportPath(ID exporter.ResourceInput) string {
	sanitizedType := strings.ReplaceAll(string(ID.Type), "/", "_")
	sanitizedService := strings.ReplaceAll(ID.ServiceName, "/", "_")

	return filepath.Join(p.workDir, fmt.Sprintf("%s_%s.tf", sanitizedType, sanitizedService))
}
