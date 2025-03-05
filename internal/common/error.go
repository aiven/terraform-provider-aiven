package common

import (
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
)

// IsCritical returns true if the given error is critical
func IsCritical(err error) bool {
	return !(err == nil || aiven.IsNotFound(err) || avngen.IsNotFound(err))
}

const (
	// SummaryResourceDeprecated is the error summary for when a resource is deprecated.
	SummaryResourceDeprecated = "Resource Deprecated"
)

// ResourceDeprecatedError creates an error with a deprecation message
func ResourceDeprecatedError(migrationMessage string) error {
	return fmt.Errorf("%s: %s",
		SummaryResourceDeprecated,
		migrationMessage,
	)
}
