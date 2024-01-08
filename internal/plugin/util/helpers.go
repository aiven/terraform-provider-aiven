package util

import "strings"

// Ref is a helper function that returns a pointer to the value passed in.
func Ref[T any](v T) *T {
	return &v
}

// Deref is a helper function that dereferences any pointer type and returns the value.
func Deref[T any](p *T) T {
	var result T

	if p != nil {
		result = *p
	}

	return result
}

// ComposeID is a helper function that composes an ID from the parts passed in.
func ComposeID(parts ...string) string {
	return strings.Join(parts, "/")
}

// BetaDescription is a helper function that returns a description for beta resources.
func BetaDescription(description string) string {
	return description + " Please note that this resource is in beta and may change without notice. " +
		"To use this resource, please set the PROVIDER_AIVEN_ENABLE_BETA environment variable."
}
