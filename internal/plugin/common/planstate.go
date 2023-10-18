// Package util is the package that contains all the utility functions in the provider.
package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ PlanState = &tfsdk.Plan{}
	_ PlanState = &tfsdk.State{}
)

// ConfigToModel is a helper function that calls the Get method on the source, stores the result in the model and
// appends the diagnostics to the diagnostics passed as argument.
func ConfigToModel(ctx context.Context, source *tfsdk.Config, model any, diagnostics *diag.Diagnostics) bool {
	diags := source.Get(ctx, model)

	diagnostics.Append(diags...)

	return !diagnostics.HasError()
}

// PlanState is the interface that defines the Get method, which is implemented by the Plan and the State structs.
type PlanState interface {
	// Get is a method that gets the value from the Plan or the State and stores it in the target.
	Get(ctx context.Context, target any) diag.Diagnostics
	// Set is a method that sets the value in the Plan or the State from the provided value.
	Set(ctx context.Context, val any) diag.Diagnostics
}

// PlanStateToModel is a helper function that calls the Get method on the source, stores the result in the model and
// appends the diagnostics to the diagnostics passed as argument.
func PlanStateToModel(ctx context.Context, source PlanState, model any, diagnostics *diag.Diagnostics) bool {
	diags := source.Get(ctx, model)

	diagnostics.Append(diags...)

	return !diagnostics.HasError()
}

// ModelToPlanState is a helper function that calls the Set method on the target, sets the value from the provided
// model and appends the diagnostics to the diagnostics passed as argument.
func ModelToPlanState(ctx context.Context, model any, target PlanState, diagnostics *diag.Diagnostics) bool {
	diags := target.Set(ctx, model)

	diagnostics.Append(diags...)

	return !diagnostics.HasError()
}
