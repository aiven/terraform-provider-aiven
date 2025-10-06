package adapter

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Model implements resource or datasource model with the shared fields model.
type Model[T any] interface {
	// SharedModel returns the shared fields model between resource and datasource.
	SharedModel() *T
}

// newModel returns a new instance of the Model.
type newModel[T any] func() Model[T]

// DatView datasource view interface
type DatView[T any] interface {
	// Configure sets the client for the view and potentially other dependencies, like logging.
	Configure(client avngen.Client)
	Read(ctx context.Context, state *T) diag.Diagnostics
}

// ResView represents a resource view that handles CRUD operations. It uses three main state objects:
//
// - state: The current state of the resource containing all values (required, optional, computed)
// - config: The raw configuration from the user which may contain unknown values due to interpolation
// - plan: The planned final state, combining user-defined values from "config" with computed/optional values from "state"
//
// Create and Update operations typically use plan as it represents the desired end state.
// "config" is used when we need to check what values the user explicitly defined.
//
// For optional+computed attributes with UseStateForUnknown:
// When a user removes a value from "config", the value persists in "state" and appears in "plan",
// allowing the attribute to retain its last known value.
type ResView[T any] interface {
	DatView[T]
	Create(ctx context.Context, plan *T) diag.Diagnostics
	Update(ctx context.Context, plan, state, config *T) diag.Diagnostics
	Delete(ctx context.Context, state *T) diag.Diagnostics
}

// DatConfigValidators implements datasource.DataSourceWithConfigValidators.
// It renames the method, so it can be used with ResConfigValidators without collisions.
type DatConfigValidators[T any] interface {
	DatConfigValidators(ctx context.Context) []datasource.ConfigValidator
}

// ResConfigValidators implements resource.ResourceWithConfigValidators.
// It renames the method, so it can be used with DatConfigValidators without collisions.
// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#configvalidators-method
type ResConfigValidators[T any] interface {
	ResConfigValidators(ctx context.Context) []resource.ConfigValidator
}

// ResValidateConfig implements resource.ResourceWithValidateConfig.
// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#validateconfig-method
type ResValidateConfig[T any] interface {
	ResValidateConfig(ctx context.Context, config *T) diag.Diagnostics
}

// View base view that contains the client and potentially other dependencies
type View struct {
	Client avngen.Client
}

func (v *View) Configure(client avngen.Client) {
	v.Client = client
}
