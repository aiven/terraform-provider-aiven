package errmsg

import (
	"context"
	"errors"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RetryDiags detects DiagError in diagnostics and retries it.
// Use FromError to create a DiagError from diag.Diagnostic.
func RetryDiags(ctx context.Context, retryableFunc func() diag.Diagnostics, opts ...retry.Option) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = retry.Do(func() error {
		diags = retryableFunc()
		for _, d := range diags {
			if o, ok := d.(DiagError); ok {
				// Use retry.RetryIf to control which error to retry.
				// By default, retries all errors.
				return o.Error
			}
		}

		// No DiagError found, exit.
		return nil
	}, append(opts, retry.Context(ctx))...)
	return diags
}

func RetryIfAivenError(f func(avngen.Error) bool) retry.Option {
	return retry.RetryIf(func(err error) bool {
		var e avngen.Error
		return errors.As(err, &e) && f(e)
	})
}

func RetryIfAivenStatus(codes ...int) retry.Option {
	return RetryIfAivenError(func(e avngen.Error) bool {
		return slices.Contains(codes, e.Status)
	})
}

// WarnDiagError turns errors into warnings if they match the filter function.
func WarnDiagError(diags diag.Diagnostics, f func(error) bool) diag.Diagnostics {
	for i, d := range diags {
		if e, ok := d.(DiagError); ok && f(e.Error) {
			diags[i] = diag.NewWarningDiagnostic(d.Summary(), d.Detail())
		}
	}
	return diags
}
