package common

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
)

// IsCritical returns true if the given error is critical
func IsCritical(err error) bool {
	return err != nil && !aiven.IsNotFound(err) && !avngen.IsNotFound(err)
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

// NewNotFound creates a new not found error
// There are lots of endpoints that return a list of objects which might not contain the object we are looking for.
// In this case, we should still return 404.
func NewNotFound(msg string, args ...any) error {
	return aiven.Error{Status: http.StatusNotFound, Message: fmt.Sprintf(msg, args...)}
}

func IsNotFound(err error) bool {
	return aiven.IsNotFound(err) || avngen.IsNotFound(err)
}

func OmitNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}

// isUnknownRole checks if the database returned an error because of an unknown role
// to make deletions idempotent
func isUnknownRole(err error) bool {
	var msg string
	var oldErr aiven.Error
	var newErr avngen.Error
	switch {
	case errors.As(err, &oldErr):
		msg = oldErr.Message
	case errors.As(err, &newErr):
		msg = newErr.Message
	default:
		return false
	}
	return strings.Contains(msg, "Code: 511")
}

// IsUnknownResource is a function to handle errors that we want to treat as "Not Found"
func IsUnknownResource(err error) bool {
	return IsNotFound(err) || isUnknownRole(err)
}
