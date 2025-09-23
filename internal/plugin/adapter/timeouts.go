// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type timeoutType string

const (
	timeoutCreate  timeoutType = "create"
	timeoutRead    timeoutType = "read"
	timeoutUpdate  timeoutType = "update"
	timeoutDelete  timeoutType = "delete"
	timeoutDefault timeoutType = "default"
)

// withTimeout returns a new context with the specified timeout from the timeouts object.
// Uses schemautil.GetDefaultTimeout() value from "ldflags" as the fallback.
func withTimeout(ctx context.Context, timeouts types.Object, timeoutKey timeoutType) (context.Context, func(), diag.Diagnostics) {
	timeout, diags := getTimeoutValue(timeouts, timeoutKey, schemautil.GetDefaultTimeout())
	if diags.HasError() {
		// Return original context if timeout value is invalid to avoid nil pointer dereference.
		return ctx, nil, diags
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel, nil
}

// getTimeoutValue Retrieves the timeout value for the specified timeoutKey.
// For SDKv2 compatibility, we use custom timeouts blocks with "default" key.
func getTimeoutValue(timeouts types.Object, timeoutKey timeoutType, fallback time.Duration) (time.Duration, diag.Diagnostics) {
	values := make(map[timeoutType]time.Duration)
	for k, v := range timeouts.Attributes() {
		if v.IsNull() || v.IsUnknown() {
			continue
		}

		// Lets it fail if the type is not string, it is not a user error.
		duration, err := time.ParseDuration(v.(types.String).ValueString())
		if err != nil {
			return 0, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid Timeout Value",
				fmt.Sprintf("Failed to parse timeout value for %q: %s", timeoutKey, err.Error()),
			)}
		}
		values[timeoutType(k)] = duration
	}

	// Gets specific timeout
	if v, ok := values[timeoutKey]; ok {
		return v, nil
	}

	// Or fallback to "default" timeout
	if v, ok := values[timeoutDefault]; ok {
		return v, nil
	}

	// Or fallback to the provided fallback value
	return fallback, nil
}
