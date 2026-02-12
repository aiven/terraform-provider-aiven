package util

import (
	"os"
	"strings"
)

const (
	// errTerraformTypeAssertionFailed is an error that is returned when a Terraform type assertion fails.
	errTerraformTypeAssertionFailed = "terraform type assertion failed"
	AivenEnableBeta                 = "PROVIDER_AIVEN_ENABLE_BETA"
)

// IsBeta returns true if beta features are enabled.
// This is used in the provider schema to determine if beta features should be included.
// In case this functionality is needed in tests, please use acc.SkipIfNotBeta(t) to skip tests when beta features are not enabled.
func IsBeta() bool {
	switch strings.ToLower(os.Getenv(AivenEnableBeta)) {
	case "false", "":
		// The previous implementation allowed any "a non-zero value" to enable beta.
		// For backward compatibility, we explicitly check for "false".
		return false
	}
	return true
}

// ComposeID is a helper function that composes an ID from the parts passed in.
func ComposeID(parts ...string) string {
	return strings.Join(parts, "/")
}
