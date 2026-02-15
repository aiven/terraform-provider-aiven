package awsentity

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// NewResource returns the entity resource.
// The generated CRUD is used directly — idAttributeComposed uses
// [organization_id, custom_cloud_environment_id] which produces
// correct parameter ordering for all API calls.
func NewResource() resource.Resource {
	return adapter.NewResource(ResourceOptions)
}
