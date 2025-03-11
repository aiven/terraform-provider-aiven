package template

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
	sdkGenerator       *SDKTemplateGenerator
	frameworkGenerator *FrameworkTemplateGenerator
}

// NewStore creates a new empty template store for SDK-v2 tests
func NewStore(t testing.TB) *Store {
	return &Store{
		registry:           newTemplateRegistry(t),
		t:                  t,
		sdkGenerator:       NewSDKTemplateGenerator(),
		frameworkGenerator: NewFrameworkTemplateGenerator(),
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

	set := NewStore(t)

	// Register all resources
	for resourceType, resource := range p.ResourcesMap {
		set.registerSDKResource(resourceType, resource, ResourceKindResource)
	}

	// Register all data sources
	for resourceType, resource := range p.DataSourcesMap {
		set.registerSDKResource(resourceType, resource, ResourceKindDataSource)
	}

	// Register all framework resources and data sources
	aivenProvider := plugin.New("dev")()

	// Get framework resources
	for _, resourceFunc := range aivenProvider.Resources(context.Background()) {
		r := resourceFunc()
		var resp frameworkresource.MetadataResponse
		r.Metadata(context.Background(), frameworkresource.MetadataRequest{ProviderTypeName: "aiven"}, &resp)
		set.registerFrameworkComponent(resp.TypeName, r, ResourceKindResource)
	}

	// Get framework data sources
	for _, dataSourceFunc := range aivenProvider.DataSources(context.Background()) {
		d := dataSourceFunc()
		var resp frameworkdatasource.MetadataResponse
		d.Metadata(context.Background(), frameworkdatasource.MetadataRequest{ProviderTypeName: "aiven"}, &resp)
		set.registerFrameworkComponent(resp.TypeName, d, ResourceKindDataSource)
	}

	for resourceType, resource := range externalTemplates {
		set.AddExternalTemplate(resourceType, resource)
	}

	return set
}

// registerSDKResource handles the registration of a single resource or data source
func (s *Store) registerSDKResource(resourceType string, r *sdkschema.Resource, kind ResourceKind) {
	// Generate and register the template
	template, err := s.sdkGenerator.GenerateTemplate(r, resourceType, kind)
	if err != nil {
		s.t.Fatalf("failed to generate SDK template for %s %s: %v", kind, resourceType, err)
	}
	s.registry.mustAddTemplate(templateKey(resourceType, kind), template)
}

// registerFrameworkComponent is a generic function to handle registration of any framework component
func (s *Store) registerFrameworkComponent(resourceType string, schemaProvider interface{}, kind ResourceKind) {
	ctx := context.Background()
	var schema interface{}

	// Extract schema based on the type of provider
	switch p := schemaProvider.(type) {
	case frameworkresource.Resource:
		// Handle Framework resource
		var resp frameworkresource.SchemaResponse
		p.Schema(ctx, frameworkresource.SchemaRequest{}, &resp)
		schema = resp.Schema
	case frameworkdatasource.DataSource:
		// Handle Framework datasource
		var resp frameworkdatasource.SchemaResponse
		p.Schema(ctx, frameworkdatasource.SchemaRequest{}, &resp)
		schema = resp.Schema
	default:
		s.t.Fatalf("unsupported schema provider type: %T", schemaProvider)
		return
	}

	// Generate and register the template
	template, err := s.frameworkGenerator.GenerateTemplate(schema, resourceType, kind)
	if err != nil {
		s.t.Fatalf("failed to generate framework template for %s %s: %v", kind, resourceType, err)
	}
	s.registry.mustAddTemplate(templateKey(resourceType, kind), template)
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
