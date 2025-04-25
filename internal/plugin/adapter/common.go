package adapter

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type DataView[T any] interface {
	// Configure sets the client for the view and potentially other dependencies, like logging.
	Configure(client avngen.Client)
	Read(ctx context.Context, state *T) diag.Diagnostics
}

type View[T any] interface {
	DataView[T]
	Create(ctx context.Context, plan *T) diag.Diagnostics
	Update(ctx context.Context, plan, state *T) diag.Diagnostics
	Delete(ctx context.Context, state *T) diag.Diagnostics
}

type DataModel[T any] interface {
	// DataModel returns embedded dataModel instance
	DataModel() *T
}

type dataModelFactory[T any] func() DataModel[T]
