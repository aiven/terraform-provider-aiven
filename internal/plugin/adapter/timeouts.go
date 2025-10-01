// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	timeout, diags := getTimeoutValue(ctx, timeouts, timeoutKey, schemautil.GetDefaultTimeout())
	if diags.HasError() {
		// Return original context if timeout value is invalid to avoid nil pointer dereference.
		return ctx, nil, diags
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel, nil
}

// getTimeoutValue retrieves the timeout value in SDKv2 legacy manner (the plugin framework doesn't have "default" key):
// 1. The specific timeout value for the given key (e.g. "create", "read", "update", "delete")
// 2. The "default" timeout value if specific one not found
// 3. The provided fallback duration if no timeouts configured
// Returns the resolved duration and any validation diagnostics.
func getTimeoutValue(
	ctx context.Context,
	timeouts types.Object,
	timeoutKey timeoutType,
	fallback time.Duration,
) (time.Duration, diag.Diagnostics) {
	values := make(map[timeoutType]time.Duration)
	for k, v := range timeouts.Attributes() {
		if v.IsNull() || v.IsUnknown() {
			continue
		}

		// The schema ensures that the value is a string.
		duration, err := time.ParseDuration(v.(types.String).ValueString())
		if err != nil {
			return 0, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid Timeout Value",
				fmt.Sprintf("Failed to parse timeout value for %q: %s", timeoutKey, err.Error()),
			)}
		}
		values[timeoutType(k)] = duration
	}

	if v, ok := values[timeoutKey]; ok {
		tflog.Info(ctx, fmt.Sprintf("Using %q timeout: %s", timeoutKey, v))
		return v, nil
	}

	if v, ok := values[timeoutDefault]; ok {
		tflog.Info(ctx, fmt.Sprintf("Using %q value for %q timeout: %s", timeoutDefault, timeoutKey, v))
		return v, nil
	}

	tflog.Info(ctx, fmt.Sprintf("Using fallback timeout for %q: %s", timeoutKey, fallback))
	return fallback, nil
}
