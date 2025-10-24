package errmsg

import (
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RetryDiags detects DiagError in diagnostics and retries it.
// Use FromError to create a DiagError from diag.Diagnostic.
func RetryDiags(
	retryableFunc func() diag.Diagnostics,
	opts ...retry.Option,
) diag.Diagnostics {
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
	}, opts...)
	return diags
}
