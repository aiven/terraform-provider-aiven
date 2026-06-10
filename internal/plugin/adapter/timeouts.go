// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"context"
	"fmt"
	"time"

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
func withTimeout(ctx context.Context, d resourceDataTimeouts, timeoutKey timeoutType) (context.Context, func(), error) {
	timeout, err := getTimeoutValue(ctx, d, timeoutKey, schemautil.GetDefaultTimeout())
	if err != nil {
		// Return original context if timeout value is invalid to avoid nil pointer dereference.
		return ctx, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel, nil
}

type resourceDataTimeouts interface {
	GetOk(key string) (any, bool)
	Schema() *Schema
}

// getTimeoutValue retrieves the timeout value:
// 1. The specific timeout value for the given key (e.g. "create", "read", "update", "delete")
// 2. The legacy "default" timeout value if the schema supports it and specific one not found
// 3. The provided fallback duration if no timeouts configured
// Returns the resolved duration and any validation diagnostics.
func getTimeoutValue(
	ctx context.Context,
	d resourceDataTimeouts,
	timeoutKey timeoutType,
	fallback time.Duration,
) (time.Duration, error) {
	// Note: the timeouts block is not represented as a list of objects, hence the index is not used.
	v, ok := d.GetOk(fmt.Sprintf("timeouts.%s", timeoutKey))
	if ok {
		tflog.Info(ctx, fmt.Sprintf("Using user %q timeout: %s", timeoutKey, v))
	} else {
		if schema := d.Schema(); schema != nil {
			if timeouts, hasTimeouts := schema.Properties["timeouts"]; hasTimeouts && timeouts != nil {
				if _, hasTimeoutDefault := timeouts.Properties[string(timeoutDefault)]; hasTimeoutDefault {
					v, ok = d.GetOk(fmt.Sprintf("timeouts.%s", timeoutDefault))
					if ok {
						tflog.Info(ctx, fmt.Sprintf("Using %q value for %q timeout: %s", timeoutDefault, timeoutKey, v))
					}
				}
			}
		}
		if !ok {
			tflog.Info(ctx, fmt.Sprintf("Using fallback timeout for %q: %s", timeoutKey, fallback))
			return fallback, nil
		}
	}

	duration, err := time.ParseDuration(v.(string))
	if err != nil {
		return 0, fmt.Errorf("failed to parse timeout value for %q: %w", timeoutKey, err)
	}

	return duration, nil
}
