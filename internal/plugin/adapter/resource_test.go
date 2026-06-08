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
	const (
		fastRetryDelay = time.Microsecond
		fastAttempts   = uint(5)
	)

	// fastAdapter builds a *resourceAdapter wired for fast tests: short retry delay and a low
	// attempt cap. Per-test config (RefreshStateDelay, RefreshStateDesired, etc.) is layered on
	// top of the returned adapter's resource by the caller.
	fastAdapter := func(read func(ctx context.Context, client avngen.Client, rd ResourceData) error) *resourceAdapter {
		return &resourceAdapter{
			resource: ResourceOptions{
				Read:                    read,
				refreshStateRetryDelay:  fastRetryDelay,
				refreshStateMaxAttempts: fastAttempts,
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

	t.Run("returns ErrRefreshStateDesired after exhausting retries", func(t *testing.T) {
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
			return nil
		})
		a.resource.RefreshStateDesired = map[string]string{"status": "ACTIVE"}
		a.resource.refreshStateMaxAttempts = 3

		err = a.refreshState(t.Context(), rd)

		require.ErrorIs(t, err, ErrRefreshStateDesired)
		require.Equal(t, 3, attempts)
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
