package adapter

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Model implements resource or datasource model with a shared fields model.
type Model[T any] interface {
	// SharedModel returns shared fields model between resource and datasource.
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

// ResView resource view interface
type ResView[T any] interface {
	DatView[T]
	Create(ctx context.Context, plan *T) diag.Diagnostics
	Update(ctx context.Context, plan, state *T) diag.Diagnostics
	Delete(ctx context.Context, state *T) diag.Diagnostics
}

// DatViewValidators implements datasource.DataSourceWithConfigValidators.
// It renames the method, so it can be used with ResViewValidators without collisions.
type DatViewValidators[T any] interface {
	DatView[T]
	DatValidators(ctx context.Context) []datasource.ConfigValidator
}

// ResViewValidators implements resource.ResourceWithConfigValidators.
// It renames the method, so it can be used with DatViewValidators without collisions.
type ResViewValidators[T any] interface {
	ResView[T]
	ResValidators(ctx context.Context) []resource.ConfigValidator
}

// View base view that contains the client and potentially other dependencies
type View struct {
	Client avngen.Client
}

func (v *View) Configure(client avngen.Client) {
	v.Client = client
}
