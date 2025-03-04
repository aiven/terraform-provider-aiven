package template

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

var (
	globalStore *Store
	stSetOnce   sync.Once
	envLock     sync.Mutex
)

// Store represents a set of related resources and their templates
type Store struct {
	registry           *registry
	t                  testing.TB
	generator          *TextTemplateGenerator
	frameworkGenerator *FrameworkTextTemplateGenerator
}

// NewSDKStore creates a new empty template store for SDK-v2 tests
func NewSDKStore(t testing.TB) *Store {
	return &Store{
		registry:           newTemplateRegistry(t),
		t:                  t,
		generator:          NewSchemaTemplateGenerator(),
		frameworkGenerator: NewFrameworkSchemaTemplateGenerator(),
	}
}

// NewFrameworkStore creates a new store specifically for framework resources
func NewFrameworkStore(t testing.TB) *Store {
	return &Store{
		registry:           newTemplateRegistry(t),
		t:                  t,
		generator:          NewSchemaTemplateGenerator(),
		frameworkGenerator: NewFrameworkSchemaTemplateGenerator(),
	}
}

// InitializeTemplateStore initializes a singleton instance of Store with all provider resources and templates pre-registered
func InitializeTemplateStore(t testing.TB) *Store {
	stSetOnce.Do(func() {
		globalStore = initTemplateStore(t)
	})

	globalStore.t = t // Update testing.TB for the current test context

	return globalStore
}

// AddTemplate adds a custom template to the template set
func (s *Store) AddTemplate(name, templateStr string) *Store {
	s.registry.mustAddTemplate(name, templateStr)

	return s
}

// AddExternalTemplate adds an external template to the store with the appropriate prefix
func (s *Store) AddExternalTemplate(name, templateStr string) *Store {
	// Determine the template key based on the name prefix
	var tk string

	if strings.HasPrefix(name, "data.") {
		// Data source template
		dataSourceName := strings.TrimPrefix(name, "data.")
		tk = fmt.Sprintf("data.%s", dataSourceName)
	} else if strings.HasPrefix(name, "provider.") {
		// Provider template - keep name as is
		tk = name
	} else if strings.HasSuffix(name, "_provider") {
		// Provider template with _provider suffix - keep name as is
		tk = name
	} else {
		// Default to resource template
		tk = fmt.Sprintf("resource.%s", name)
	}

	s.registry.mustAddTemplate(tk, templateStr)
	return s
}

// NewBuilder creates a new composition builder
func (s *Store) NewBuilder() *CompositionBuilder {
	return &CompositionBuilder{
		registry:     s.registry,
		compositions: make([]compositionEntry, 0),
	}
}

// initTemplateStore initializes a new template set with all provider resources
func initTemplateStore(t testing.TB) *Store {
	// Acquire lock before modifying environment variables
	envLock.Lock()

	// Create a temporary test context with beta flag set
	originalValue := os.Getenv(util.AivenEnableBeta)

	// Set up deferred unlock and environment restoration
	defer func() {
		// Restore original value when function exits
		if originalValue == "" {
			_ = os.Unsetenv(util.AivenEnableBeta)
		} else {
			_ = os.Setenv("PROVIDER_AIVEN_ENABLE_BETA", originalValue)
		}

		// Release the lock
		envLock.Unlock()
	}()

	// Set beta flag for initializing the provider with all resources
	if err := os.Setenv(util.AivenEnableBeta, "true"); err != nil {
		t.Fatalf("failed to set beta flag: %v", err)
	}

	p, err := provider.Provider("dev")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}

	set := NewSDKStore(t)

	// Register all resources
	for resourceType, resource := range p.ResourcesMap {
		set.registerResource(resourceType, resource, ResourceKindResource)
	}

	// Register all data sources
	for resourceType, resource := range p.DataSourcesMap {
		set.registerResource(resourceType, resource, ResourceKindDataSource)
	}

	// Register all framework resources and data sources
	aivenProvider := plugin.New("dev")()

	// Get framework resources
	for _, resourceFunc := range aivenProvider.Resources(context.Background()) {
		r := resourceFunc()
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "aiven"}, &resp)
		set.registerFrameworkResource(resp.TypeName, r)
	}

	// Get framework data sources
	for _, dataSourceFunc := range aivenProvider.DataSources(context.Background()) {
		d := dataSourceFunc()
		var resp datasource.MetadataResponse
		d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "aiven"}, &resp)
		set.registerFrameworkDataSource(resp.TypeName, d)
	}

	for resourceType, resource := range externalTemplates {
		set.AddExternalTemplate(resourceType, resource)
	}

	return set
}

// registerResource handles the registration of a single resource or data source
func (s *Store) registerResource(resourceType string, r *schema.Resource, kind ResourceKind) {
	// Generate and register the template
	template := s.generator.GenerateTemplate(r, resourceType, kind)
	s.registry.mustAddTemplate(templateKey(resourceType, kind), template)
}

// registerFrameworkResource handles the registration of a framework resource
func (s *Store) registerFrameworkResource(resourceType string, r resource.Resource) {
	// Get the schema from the resource
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// Generate and register the template
	template := s.frameworkGenerator.GenerateTemplate(resp.Schema, resourceType, ResourceKindResource)
	s.registry.mustAddTemplate(templateKey(resourceType, ResourceKindResource), template)
}

// registerFrameworkDataSource handles the registration of a framework data source
func (s *Store) registerFrameworkDataSource(resourceType string, d datasource.DataSource) {
	// Get the schema from the data source
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// Generate and register the template
	template := s.frameworkGenerator.GenerateTemplate(resp.Schema, resourceType, ResourceKindDataSource)
	s.registry.mustAddTemplate(templateKey(resourceType, ResourceKindDataSource), template)
}

// templateKey generates a unique template key based on resource type and kind
func templateKey(resourceType string, kind ResourceKind) string {
	switch kind {
	case ResourceKindResource:
		return fmt.Sprintf("resource.%s", resourceType)
	case ResourceKindDataSource:
		return fmt.Sprintf("data.%s", resourceType)
	default:
		return resourceType
	}
}
