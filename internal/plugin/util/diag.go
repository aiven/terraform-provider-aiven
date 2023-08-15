// Package util is the package that contains all the utility functions in the provider.
package util

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

<<<<<<< HEAD:internal/plugin/util/diag.go
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
=======
	"github.com/aiven/terraform-provider-aiven/internal/provider/errmsg"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283)):internal/provider/util/diag.go
)

// DiagErrorUnexpectedProviderDataType is a function that adds an unexpected provider data type error to the
// diagnostics and returns it. It is used in the Configure method of the resource structs.
func DiagErrorUnexpectedProviderDataType(diagnostics diag.Diagnostics, providerData any) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryUnexpectedProviderDataType,
		fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, providerData),
	)

	return diagnostics
}

// DiagErrorCreatingResource is a function that adds a resource creation error to the diagnostics and returns it.
// It is used in the Create method of the resource structs.
func DiagErrorCreatingResource(diagnostics diag.Diagnostics, typenameable TypeNameable, err error) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryErrorCreatingResource,
		fmt.Sprintf(errmsg.DetailErrorCreatingResource, typenameable.TypeName(), err.Error()),
	)

	return diagnostics
}

// DiagErrorReadingResource is a function that adds a resource reading error to the diagnostics and returns it.
// It is used in the Read method of the resource structs.
func DiagErrorReadingResource(diagnostics diag.Diagnostics, typenameable TypeNameable, err error) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryErrorReadingResource,
		fmt.Sprintf(errmsg.DetailErrorReadingResource, typenameable.TypeName(), err.Error()),
	)

	return diagnostics
}

// DiagErrorUpdatingResource is a function that adds a resource updating error to the diagnostics and returns it.
// It is used in the Update method of the resource structs.
func DiagErrorUpdatingResource(diagnostics diag.Diagnostics, typenameable TypeNameable, err error) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryErrorUpdatingResource,
		fmt.Sprintf(errmsg.DetailErrorUpdatingResource, typenameable.TypeName(), err.Error()),
	)

	return diagnostics
}

// DiagErrorDeletingResource is a function that adds a resource deleting error to the diagnostics and returns it.
// It is used in the Delete method of the resource structs.
func DiagErrorDeletingResource(diagnostics diag.Diagnostics, typenameable TypeNameable, err error) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryErrorDeletingResource,
		fmt.Sprintf(errmsg.DetailErrorDeletingResource, typenameable.TypeName(), err.Error()),
	)

	return diagnostics
}

// DiagErrorReadingDataSource is a function that adds a data source reading error to the diagnostics and returns it.
// It is used in the Read method of the data source structs.
func DiagErrorReadingDataSource(diagnostics diag.Diagnostics, typenameable TypeNameable, err error) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryErrorReadingDataSource,
		fmt.Sprintf(errmsg.DetailErrorReadingDataSource, typenameable.TypeName(), err.Error()),
	)

	return diagnostics
}

// DiagDuplicateFoundByName is a function that adds a duplicate found by name error to the diagnostics and returns it.
func DiagDuplicateFoundByName(diagnostics diag.Diagnostics, name string) diag.Diagnostics {
	diagnostics.AddError(
		errmsg.SummaryDuplicateFoundByName, fmt.Sprintf(errmsg.DetailDuplicateFoundByName, name),
	)

	return diagnostics
}
