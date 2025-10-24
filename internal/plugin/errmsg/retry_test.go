package errmsg

import (
	"net/http"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
)

func TestRetryDiags(t *testing.T) {
	testcases := []struct {
		name          string
		diags         diag.Diagnostics
		expectRetried bool
		opts          []retry.Option
	}{
		{
			name:          "no error",
			expectRetried: false,
		},
		{
			name:          "doesn't retry unknown errors",
			expectRetried: false,
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test", "test"),
			},
		},
		{
			name:          "retries 404",
			expectRetried: true,
			diags: diag.Diagnostics{
				FromError("test", avngen.Error{Status: http.StatusNotFound}),
			},
			opts: []retry.Option{
				retry.RetryIf(avngen.IsNotFound),
			},
		},
		{
			name:          "doesn't retry 400",
			expectRetried: false,
			diags: diag.Diagnostics{
				FromError("test", avngen.Error{Status: http.StatusBadRequest}),
			},
			opts: []retry.Option{
				retry.RetryIf(avngen.IsNotFound),
			},
		},
		{
			name:          "retries 400, because no RetryIf specified",
			expectRetried: true,
			diags: diag.Diagnostics{
				FromError("test", avngen.Error{Status: http.StatusBadRequest}),
			},
		},
	}

	// We set maxAttempts so test doesn't run forever
	const maxAttempts = 2

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts = append(
				tc.opts,
				retry.Attempts(maxAttempts),
				retry.Delay(time.Millisecond),
			)

			attempts := 0
			_ = RetryDiags(func() diag.Diagnostics {
				attempts++
				return tc.diags
			}, tc.opts...)

			assert.Equal(t, tc.expectRetried, attempts == maxAttempts)
		})
	}
}
