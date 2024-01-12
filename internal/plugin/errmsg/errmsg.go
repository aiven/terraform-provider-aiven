// Package errmsg is the package that contains all the error messages in the provider.
package errmsg

// All error messages in the provider should be defined in this package.
// This is to ensure that all error messages are consistent and not duplicated, and follow the same style.
// This is also to ensure that all error messages are defined in one place, and are easy to find and update:
//
//	- If you need to add a new error message, please add it to this package.
//	- If you need to update an existing error message, please update it in this package.
//	- If you need to use an error message, please import it from this package.

// Below is the list of error summaries that are used in the provider.
// The error summaries are used to identify the error type, and are used in the error handling.
// The error summaries should use title case heading and SHOULD NOT end with a period.
// The error summaries SHOULD NOT contain placeholders for values that are not known at the time of writing.
//
//	See APA style guide for more information:
//	  https://apastyle.apa.org/style-grammar-guidelines/capitalization/title-case.
const (
	// SummaryUnexpectedError is the error summary for when an unexpected error occurs.
	// This is the default error summary that is used when no other error summary is applicable.
	// It is advised to use a more specific error summary if possible.
	SummaryUnexpectedError = "Unexpected Error"

	// SummaryTokenMissing is the error summary for when a token is missing.
	SummaryTokenMissing = "Token Missing"

	// SummaryConstructingClient is the error summary for when a client cannot be constructed.
	SummaryConstructingClient = "Constructing Client Failed"

	// SummaryUnexpectedProviderDataType is the error summary for when the provider data type is unexpected.
	SummaryUnexpectedProviderDataType = "Unexpected Provider Data Type"

	// SummaryErrorCreatingResource is the error summary for when a resource cannot be created.
	SummaryErrorCreatingResource = "Error Creating Resource"

	// SummaryErrorReadingResource is the error summary for when a resource cannot be read.
	SummaryErrorReadingResource = "Error Reading Resource"

	// SummaryErrorUpdatingResource is the error summary for when a resource cannot be updated.
	SummaryErrorUpdatingResource = "Error Updating Resource"

	// SummaryErrorDeletingResource is the error summary for when a resource cannot be deleted.
	SummaryErrorDeletingResource = "Error Deleting Resource"

	// SummaryErrorImportingResource is the error summary for when a resource cannot be imported.
	SummaryErrorImportingResource = "Error Importing Resource"

	// SummaryDuplicateFoundByName is the error summary for when a duplicate resource is found by name.
	SummaryDuplicateFoundByName = "Duplicate Found By Name"

	// SummaryErrorReadingDataSource is the error summary for when a data source cannot be read.
	SummaryErrorReadingDataSource = "Error Reading Data Source"
)

// Below is the list of detailed error messages that are used in the provider.
// The detailed error messages are used to provide more information about the error.
// The detailed error messages should use sentence case and should end with a period.
// The detailed error messages may contain placeholders for values that are not known at the time of writing.
//
//	See APA style guide for more information:
//	  https://apastyle.apa.org/style-grammar-guidelines/capitalization/sentence-case.
var (
	// DetailUnexpectedError is the detailed error message for when an unexpected error occurs.
	// This is the default detailed error message that is used when no other detailed error message is applicable.
	// It is advised to use a more specific detailed error message if possible.
	DetailUnexpectedError = "An unexpected error occurred: %s."

	// DetailTokenMissing is the detailed error message for when a token is missing.
	DetailTokenMissing = "Aiven API token was not set in the provider configuration or in the AIVEN_TOKEN " +
		"environment variable."

	// DetailUnexpectedProviderDataType is the detailed error message for when the provider data type is unexpected.
	DetailUnexpectedProviderDataType = "Expected *aiven.Client, got: %T. Please report this issue to the " +
		"provider developers."

	// DetailErrorCreatingResource is the detailed error message for when a resource cannot be created.
	DetailErrorCreatingResource = "An unexpected error occurred while creating the resource (%s): %s."

	// DetailErrorReadingResource is the detailed error message for when a resource cannot be read.
	DetailErrorReadingResource = "An unexpected error occurred while reading the resource (%s): %s."

	// DetailErrorUpdatingResource is the detailed error message for when a resource cannot be updated.
	DetailErrorUpdatingResource = "An unexpected error occurred while updating the resource (%s): %s."

	// DetailErrorUpdatingResourceNotSupported is the detailed error message for when a resource cannot be updated.
	DetailErrorUpdatingResourceNotSupported = "Updating the resource (%s) is not supported."

	// DetailErrorDeletingResource is the detailed error message for when a resource cannot be deleted.
	DetailErrorDeletingResource = "An unexpected error occurred while deleting the resource (%s): %s."

	// DetailErrorImportingResourceNotSupported is the detailed error message for when a resource cannot be imported
	// because it is not supported.
	DetailErrorImportingResourceNotSupported = "Importing the resource (%s) is not supported."

	// DetailDuplicateFoundByName is the detailed error message for when a duplicate resource is found by name.
	DetailDuplicateFoundByName = "Multiple resources with the same name (%s) were found. Please use the ID to " +
		"uniquely identify the resource."

	// DetailErrorReadingDataSource is the detailed error message for when a data source cannot be read.
	DetailErrorReadingDataSource = "An unexpected error occurred while reading the data source (%s): %s."
)

// Below is the list of classic Go-style error messages that are used in the provider.
// The classic Go-style error messages are used to provide more information about the error.
// The classic Go-style error messages should start with a lowercase letter and SHOULD NOT end with a period.
// The classic Go-style error messages may contain placeholders for values that are not known at the time of writing.
//
//	See Go error handling for more information:
//	  https://blog.golang.org/error-handling-and-go.
var (
	// ResourceNotFound is the error message for when a resource cannot be found.
	// This error is intended to be used in acceptance tests.
	ResourceNotFound = "resource not found: %s"

	// AivenResourceNotFound is the error message for when an Aiven resource cannot be found.
	AivenResourceNotFound = "aiven resource %s with compound ID %s not found"

	// UnableToSetValueFrom is the error message for when a Set cannot be created from a value.
	UnableToSetValueFrom = "unable to set value from %v"
)
