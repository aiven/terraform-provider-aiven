package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestResourceAdapter_refreshState(t *testing.T) {
	// Use short retry delays so retry-go's exponential backoff stays in the microsecond range.
	const fastRetryDelay = time.Microsecond

	// fastAdapter builds a *resourceAdapter wired for fast tests: short retry delay. Per-test config
	// (RefreshStateDelay, RefreshStateDesired, etc.) is layered on top of the returned adapter's
	// resource by the caller.
	fastAdapter := func(read func(ctx context.Context, client avngen.Client, rd ResourceData) error) *resourceAdapter {
		return &resourceAdapter{
			resource: ResourceOptions{
				Read:                   read,
				refreshStateRetryDelay: fastRetryDelay,
			},
		}
	}

	// statusSchema is a minimal schema used by tests that exercise RefreshStateDesired.
	statusSchema := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"status": {Type: SchemaTypeString},
		},
	}

	warning := func(summary string) diag.Diagnostic {
		return diag.NewWarningDiagnostic(summary, "detail")
	}

	t.Run("keeps outer warnings and merges only last try warnings", func(t *testing.T) {
		var target diag.Diagnostics

		outerCtx, drainWarnings := withWarnings(t.Context(), &target)
		AddWarnings(outerCtx, warning("outer-1"), warning("outer-2"))

		outerCollector, ok := outerCtx.Value(warningCollectorKey{}).(*warningCollector)
		require.True(t, ok)
		require.NotNil(t, outerCollector)

		warningsByTry := [][]diag.Diagnostic{
			{warning("try-1-a"), warning("try-1-b")},
			{warning("try-2-a"), warning("try-2-b")},
		}

		attempts := 0
		a := &resourceAdapter{
			resource: ResourceOptions{
				Read: func(ctx context.Context, _ avngen.Client, _ ResourceData) error {
					attempts++
					AddWarnings(ctx, warningsByTry[attempts-1]...)

					if attempts == 1 {
						return avngen.Error{Status: http.StatusNotFound, Message: "not ready"}
					}

					return nil
				},
			},
		}

		err := a.refreshState(outerCtx, nil)
		drainWarnings()

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
		require.Equal(t, 4, target.WarningsCount())
		require.True(t, target.Contains(warning("outer-1")))
		require.True(t, target.Contains(warning("outer-2")))
		require.True(t, target.Contains(warning("try-2-a")))
		require.True(t, target.Contains(warning("try-2-b")))
		require.False(t, target.Contains(warning("try-1-a")))
		require.False(t, target.Contains(warning("try-1-b")))
	})

	t.Run("returns nil when read succeeds on the first try", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			return nil
		})

		err := a.refreshState(t.Context(), nil)

		require.NoError(t, err)
		require.Equal(t, 1, attempts)
	})

	t.Run("does not retry on a non-retryable error", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("boom")
		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			return boom
		})

		err := a.refreshState(t.Context(), nil)

		require.ErrorIs(t, err, boom)
		require.Equal(t, 1, attempts, "non-retryable errors must not trigger retries")
	})

	t.Run("retries on 404 then succeeds", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			if attempts < 3 {
				return avngen.Error{Status: http.StatusNotFound, Message: "not ready"}
			}
			return nil
		})

		err := a.refreshState(t.Context(), nil)

		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("retries on wrapped ErrNotFound then succeeds", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			if attempts == 1 {
				return fmt.Errorf("lookup: %w", ErrNotFound)
			}
			return nil
		})

		err := a.refreshState(t.Context(), nil)

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
	})

	t.Run("retries on 403 then succeeds", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			if attempts == 1 {
				return avngen.Error{Status: http.StatusForbidden, Message: "eventual consistency"}
			}
			return nil
		})

		err := a.refreshState(t.Context(), nil)

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
	})

	t.Run("retries when desired attribute does not match, then succeeds", func(t *testing.T) {
		t.Parallel()

		// rd starts with status "PENDING" and flips to "ACTIVE" after the first Read.
		rd, err := NewResourceData(
			statusSchema,
			nil,
			WithTestState(map[string]any{"status": "PENDING"}),
		)
		require.NoError(t, err)

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, rd ResourceData) error {
			attempts++
			if attempts >= 2 {
				require.NoError(t, rd.Set("status", "ACTIVE"))
			}
			return nil
		})
		a.resource.RefreshStateDesired = map[string]string{"status": "ACTIVE"}

		err = a.refreshState(t.Context(), rd)

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
		require.Equal(t, "ACTIVE", rd.Get("status"))
	})

	t.Run("retries until the context deadline when the desired state is never reached", func(t *testing.T) {
		t.Parallel()

		rd, err := NewResourceData(
			statusSchema,
			nil,
			WithTestState(map[string]any{"status": "PENDING"}),
		)
		require.NoError(t, err)

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			return nil // Never transitions to ACTIVE, so the poll runs until ctx is cancelled.
		})
		a.resource.RefreshStateDesired = map[string]string{"status": "ACTIVE"}

		// The fixed retry delay is fastRetryDelay (1µs), so a 300ms window comfortably allows many
		// attempts before the deadline.
		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		err = a.refreshState(ctx, rd)

		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Greater(t, attempts, 1, "must keep polling until the deadline rather than giving up early")
	})

	t.Run("returns ctx error when RefreshStateDelay is interrupted by ctx cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			return nil
		})
		a.resource.RefreshStateDelay = time.Hour

		err := a.refreshState(ctx, nil)

		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 0, attempts, "read must not be called when ctx is cancelled during RefreshStateDelay")
	})

	t.Run("applies RefreshStateDelay before the first attempt", func(t *testing.T) {
		t.Parallel()

		const refreshStateDelay = 25 * time.Millisecond

		attempts := 0
		a := fastAdapter(func(_ context.Context, _ avngen.Client, _ ResourceData) error {
			attempts++
			return nil
		})
		a.resource.RefreshStateDelay = refreshStateDelay

		start := time.Now()
		err := a.refreshState(t.Context(), nil)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Equal(t, 1, attempts)
		require.GreaterOrEqual(t, elapsed, refreshStateDelay)
	})
}

func TestResourceAdapter_deleteState(t *testing.T) {
	const fastRetryDelay = time.Microsecond

	// fastAdapter builds a *resourceAdapter wired for fast tests, with the delete poller enabled.
	// Callers layer Desired on top via a.resource.DeleteState and provide the per-test delete/read
	// closures.
	fastAdapter := func(
		del func(ctx context.Context, client avngen.Client, rd ResourceData) error,
		read func(ctx context.Context, client avngen.Client, rd ResourceData) error,
	) *resourceAdapter {
		return &resourceAdapter{
			resource: ResourceOptions{
				Delete: del,
				Read:   read,
				DeleteState: &DeleteStateOptions{
					retryDelay: fastRetryDelay,
				},
			},
		}
	}

	statusSchema := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"status": {Type: SchemaTypeString},
		},
	}

	warning := func(summary string) diag.Diagnostic {
		return diag.NewWarningDiagnostic(summary, "detail")
	}

	t.Run("deletes once then surfaces the 404 when read reports the resource gone", func(t *testing.T) {
		t.Parallel()

		deletes, reads := 0, 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				return nil
			},
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				reads++
				return avngen.Error{Status: http.StatusNotFound}
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		// A 404 bubbles up; Delete's outer guard ignores it as a successful deletion.
		require.True(t, IsNotFound(a.deleteState(t.Context(), nil)))
		require.Equal(t, 1, deletes)
		require.Equal(t, 1, reads)
	})

	t.Run("polls read until the desired state is reached", func(t *testing.T) {
		t.Parallel()

		rd, err := NewResourceData(statusSchema, nil, WithTestState(map[string]any{"status": "DELETING"}))
		require.NoError(t, err)

		deletes, reads := 0, 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				return nil
			},
			func(_ context.Context, _ avngen.Client, rd ResourceData) error {
				reads++
				if reads >= 3 {
					require.NoError(t, rd.Set("status", "DELETED"))
				}
				return nil
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		require.NoError(t, a.deleteState(t.Context(), rd))
		require.Equal(t, 3, reads)
		require.Equal(t, 1, deletes, "delete must be issued once after it succeeds; only read polls after")
	})

	t.Run("not-found-only polls read until the resource is gone", func(t *testing.T) {
		t.Parallel()

		reads := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error { return nil },
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				reads++
				if reads >= 3 {
					return avngen.Error{Status: http.StatusNotFound}
				}
				return nil // Still present; no attribute conditions, so keep polling until 404.
			},
		)
		// fastAdapter has no Desired conditions, so a 404 is the only terminal.

		require.True(t, IsNotFound(a.deleteState(t.Context(), nil)))
		require.Equal(t, 3, reads)
	})

	t.Run("re-issues delete every attempt, ignoring its error, until the desired state", func(t *testing.T) {
		t.Parallel()

		rd, err := NewResourceData(statusSchema, nil, WithTestState(map[string]any{"status": "DELETING"}))
		require.NoError(t, err)

		deletes := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				// Delete keeps failing with a conflict; the error is ignored and Delete is re-issued.
				return avngen.Error{Status: http.StatusConflict, Message: "still in use"}
			},
			func(_ context.Context, _ avngen.Client, rd ResourceData) error {
				if deletes >= 3 {
					return rd.Set("status", "DELETED")
				}
				return nil
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		require.NoError(t, a.deleteState(t.Context(), rd))
		require.Equal(t, 3, deletes, "delete must be re-issued on every attempt, its error ignored")
	})

	t.Run("ignores delete errors and completes once read reports the resource gone", func(t *testing.T) {
		t.Parallel()

		deletes, reads := 0, 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				return avngen.Error{Status: http.StatusForbidden, Message: "forbidden"}
			},
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				reads++
				if reads >= 2 {
					return avngen.Error{Status: http.StatusNotFound}
				}
				return nil
			},
		)
		// No Desired conditions, so a 404 is the only terminal.

		require.True(t, IsNotFound(a.deleteState(t.Context(), nil)))
		require.Equal(t, 2, deletes, "delete errors are ignored and delete is re-issued each attempt")
	})

	t.Run("stops re-issuing delete once it succeeds after earlier conflicts", func(t *testing.T) {
		t.Parallel()

		rd, err := NewResourceData(statusSchema, nil, WithTestState(map[string]any{"status": "DELETING"}))
		require.NoError(t, err)

		deletes := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				if deletes < 3 {
					// Conflicts while dependents detach; retried until it goes through.
					return avngen.Error{Status: http.StatusConflict, Message: "still in use"}
				}
				return nil // Third call succeeds; Delete must not be issued again.
			},
			func(_ context.Context, _ avngen.Client, rd ResourceData) error {
				// Terminal state is reached only after Delete has succeeded.
				if deletes >= 3 {
					return rd.Set("status", "DELETED")
				}
				return nil
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		require.NoError(t, a.deleteState(t.Context(), rd))
		require.Equal(t, 3, deletes, "delete is re-issued only until it succeeds, then read polls alone")
	})

	t.Run("clears the delete error once delete succeeds", func(t *testing.T) {
		t.Parallel()

		deletes := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				deletes++
				if deletes == 1 {
					return avngen.Error{Status: http.StatusConflict, Message: "still in use"}
				}
				return nil // Succeeds from the second call onward.
			},
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				return nil // Resource never disappears, so the poll runs until the deadline.
			},
		)
		// No Desired conditions, so the only terminal is a 404 that never comes.

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err := a.deleteState(ctx, nil)

		// Delete succeeded after the first conflict, so the stale conflict must not be surfaced:
		// only the context deadline remains.
		require.ErrorIs(t, err, context.DeadlineExceeded)
		var apiErr avngen.Error
		require.NotErrorAs(t, err, &apiErr, "delete error must be cleared once delete succeeds")
	})

	t.Run("retries until the context deadline when the desired state is never reached", func(t *testing.T) {
		t.Parallel()

		rd, err := NewResourceData(statusSchema, nil, WithTestState(map[string]any{"status": "DELETING"}))
		require.NoError(t, err)

		reads := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error { return nil },
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				reads++
				return nil // Never transitions to DELETED, so the poll runs until ctx is cancelled.
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		// The fixed retry delay is fastRetryDelay (1µs), so a 300ms window comfortably allows many
		// polls before the deadline.
		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		err = a.deleteState(ctx, rd)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Greater(t, reads, 1, "must keep polling until the deadline rather than giving up early")
	})

	t.Run("surfaces the last delete error when the poll times out", func(t *testing.T) {
		t.Parallel()

		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				return avngen.Error{Status: http.StatusConflict, Message: "still in use"}
			},
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				return nil // Resource stays present, so the delete conflict keeps recurring until timeout.
			},
		)
		// No Desired conditions, so the only terminal is a 404 that never comes.

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err := a.deleteState(ctx, nil)

		// Both the context deadline and the underlying delete conflict must be inspectable, so the
		// user sees why the delete never completed rather than only a bare context deadline.
		require.ErrorIs(t, err, context.DeadlineExceeded)
		var apiErr avngen.Error
		require.ErrorAs(t, err, &apiErr)
		require.Equal(t, http.StatusConflict, apiErr.Status)
	})

	t.Run("does not retry on a non-retryable read error", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("boom")
		reads := 0
		a := fastAdapter(
			func(_ context.Context, _ avngen.Client, _ ResourceData) error { return nil },
			func(_ context.Context, _ avngen.Client, _ ResourceData) error {
				reads++
				return boom
			},
		)
		a.resource.DeleteState.Desired = map[string]string{"status": "DELETED"}

		err := a.deleteState(t.Context(), nil)
		require.ErrorIs(t, err, boom)
		require.Equal(t, 1, reads, "non-retryable errors must not trigger retries")
	})

	t.Run("keeps outer warnings and merges only last try warnings", func(t *testing.T) {
		var target diag.Diagnostics

		outerCtx, drainWarnings := withWarnings(t.Context(), &target)
		AddWarnings(outerCtx, warning("outer-1"))

		rd, err := NewResourceData(statusSchema, nil, WithTestState(map[string]any{"status": "DELETING"}))
		require.NoError(t, err)

		warningsByTry := [][]diag.Diagnostic{
			{warning("try-1")},
			{warning("try-2")},
		}

		reads := 0
		a := &resourceAdapter{
			resource: ResourceOptions{
				Delete: func(_ context.Context, _ avngen.Client, _ ResourceData) error { return nil },
				Read: func(ctx context.Context, _ avngen.Client, rd ResourceData) error {
					reads++
					AddWarnings(ctx, warningsByTry[reads-1]...)
					if reads >= 2 {
						return rd.Set("status", "DELETED")
					}
					return nil
				},
				DeleteState: &DeleteStateOptions{
					Desired:    map[string]string{"status": "DELETED"},
					retryDelay: fastRetryDelay,
				},
			},
		}

		err = a.deleteState(outerCtx, rd)
		drainWarnings()

		require.NoError(t, err)
		require.Equal(t, 2, reads)
		require.True(t, target.Contains(warning("outer-1")))
		require.True(t, target.Contains(warning("try-2")))
		require.False(t, target.Contains(warning("try-1")))
	})
}

func TestIsRefreshStateRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "ErrNotFound sentinel", err: ErrNotFound, want: true},
		{name: "wrapped ErrNotFound", err: fmt.Errorf("lookup: %w", ErrNotFound), want: true},
		{name: "avngen 404", err: avngen.Error{Status: http.StatusNotFound}, want: true},
		{name: "wrapped avngen 404", err: fmt.Errorf("api: %w", avngen.Error{Status: http.StatusNotFound}), want: true},
		{name: "avngen 403", err: avngen.Error{Status: http.StatusForbidden}, want: true},
		{name: "wrapped avngen 403", err: fmt.Errorf("api: %w", avngen.Error{Status: http.StatusForbidden}), want: true},
		{name: "ErrRefreshStateDesired sentinel", err: ErrRefreshStateDesired, want: true},
		{name: "wrapped ErrRefreshStateDesired", err: fmt.Errorf("mismatch: %w", ErrRefreshStateDesired), want: true},
		{name: "avngen 500", err: avngen.Error{Status: http.StatusInternalServerError}, want: false},
		{name: "avngen 400", err: avngen.Error{Status: http.StatusBadRequest}, want: false},
		{name: "unrelated error", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isRefreshStateRetryable(tt.err))
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	type fooType string
	var fooVar fooType = "foo"
	require.True(t, Equal(fooVar, "foo"))
	require.True(t, Equal(new(fooVar), "foo"))
}

func TestFindOne(t *testing.T) {
	t.Parallel()

	t.Run("returns the matching element", func(t *testing.T) {
		list := []string{"a", "b", "c"}
		got, err := FindOne(list, func(i int) bool { return list[i] == "b" })
		require.NoError(t, err)
		require.Equal(t, "b", got)
	})

	t.Run("predicate uses index, not value", func(t *testing.T) {
		list := []int{10, 20, 30}
		got, err := FindOne(list, func(i int) bool { return i == 2 })
		require.NoError(t, err)
		require.Equal(t, 30, got)
	})

	t.Run("works with struct slices", func(t *testing.T) {
		type item struct {
			id   int
			name string
		}
		list := []item{{1, "alpha"}, {2, "beta"}, {3, "gamma"}}
		got, err := FindOne(list, func(i int) bool { return list[i].id == 2 })
		require.NoError(t, err)
		require.Equal(t, item{2, "beta"}, got)
	})

	t.Run("returns ErrNotFound when no match", func(t *testing.T) {
		list := []string{"a", "b", "c"}
		got, err := FindOne(list, func(i int) bool { return list[i] == "z" })
		require.ErrorIs(t, err, ErrNotFound)
		require.Empty(t, got)
	})

	t.Run("returns ErrNotFound for empty list", func(t *testing.T) {
		got, err := FindOne([]int{}, func(_ int) bool { return true })
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, 0, got)
	})

	t.Run("returns ErrNotFound for nil list", func(t *testing.T) {
		got, err := FindOne[int](nil, func(_ int) bool { return true })
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, 0, got)
	})

	t.Run("returns ErrMultiple wrapped with count when more than one match", func(t *testing.T) {
		list := []string{"a", "b", "a", "c", "a"}
		got, err := FindOne(list, func(i int) bool { return list[i] == "a" })
		require.ErrorIs(t, err, ErrMultiple)
		require.Contains(t, err.Error(), "3")
		require.Empty(t, got)
	})

	t.Run("returns zero value of struct on error", func(t *testing.T) {
		type item struct{ id int }
		list := []item{{1}, {2}}
		got, err := FindOne(list, func(_ int) bool { return false })
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, item{}, got)
	})
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "ErrNotFound sentinel", err: ErrNotFound, want: true},
		{name: "wrapped ErrNotFound", err: fmt.Errorf("lookup failed: %w", ErrNotFound), want: true},
		{name: "avngen 404 error", err: avngen.Error{Status: http.StatusNotFound}, want: true},
		{name: "wrapped avngen 404", err: fmt.Errorf("api: %w", avngen.Error{Status: http.StatusNotFound}), want: true},
		{name: "avngen non-404 error", err: avngen.Error{Status: http.StatusInternalServerError}, want: false},
		{name: "avngen 403 error", err: avngen.Error{Status: http.StatusForbidden}, want: false},
		{name: "ErrMultiple is not not-found", err: ErrMultiple, want: false},
		{name: "unrelated error", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsNotFound(tt.err))
		})
	}
}
