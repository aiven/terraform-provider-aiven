package util

import (
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// errTerraformTypeAssertionFailed is an error that is returned when a Terraform type assertion fails.
	errTerraformTypeAssertionFailed = "terraform type assertion failed"
	AivenEnableBeta                 = "PROVIDER_AIVEN_ENABLE_BETA"
)

// IsBeta is a helper function that returns a flag that indicates whether the provider is in beta mode.
// This SHOULD NOT be used anywhere else except in the provider and acceptance tests initialization.
// In case this functionality is needed in tests, please use the acctest.CommonTestDependencies.IsBeta() function.
func IsBeta() bool {
	return os.Getenv(AivenEnableBeta) != ""
}

// ComposeID is a helper function that composes an ID from the parts passed in.
func ComposeID(parts ...string) string {
	return strings.Join(parts, "/")
}

// ValueOrDefault returns the value if not nil, otherwise returns the default value. Value is converted to type
// U if possible. If the conversion is not possible, the function panics.
//
// The null value of the Terraform type should be used for the default value unless you know what you are doing, e.g.
// types.StringNull() should be used for strings, types.BoolNull() should be used for booleans.
func ValueOrDefault[T comparable, U types.Bool | types.String](value *T, defaultValue U) U {
	if value != nil {
		switch v := any(*value).(type) {
		case bool:
			if bv, ok := any(types.BoolValue(v)).(U); ok {
				return bv
			}
		case string, time.Time:
			// time.Time is also a string in the state, so we need to handle it here.

			var str string

			switch v := v.(type) {
			case string:
				str = v
			case time.Time:
				str = v.String()
			}

			if sv, ok := any(types.StringValue(str)).(U); ok {
				return sv
			}
		default:
			panic(errTerraformTypeAssertionFailed)
		}
	}

	return defaultValue
}
