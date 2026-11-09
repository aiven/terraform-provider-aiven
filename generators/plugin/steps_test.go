package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSteps(t *testing.T) {
	cases := []struct {
		name string
		op   Operation
		err  string
	}{
		{name: "operation", op: Operation{ID: "FooDelete", Type: OperationDelete, Errors: []ErrorRule{{As: ErrorAsIgnore}}}},
		{name: "fn", op: Operation{Fn: "requireChange", Type: OperationUpdate}},
		{name: "neither", op: Operation{Type: OperationRead}, err: "operation 0: set either id or fn"},
		{name: "both", op: Operation{ID: "FooGet", Fn: "read", Type: OperationRead}, err: "operation 0: set either id or fn"},
		{name: "fn not an identifier", op: Operation{Fn: "pkg.read", Type: OperationRead}, err: `fn "pkg.read" is not a Go identifier`},
		{name: "fn with options", op: Operation{Fn: "read", Type: OperationRead, ResultKey: "foo"}, err: `fn "read" takes only type`},
		{
			name: "unknown as",
			op:   Operation{ID: "FooGet", Type: OperationRead, Errors: []ErrorRule{{As: "skip"}}},
			err:  "FooGet: errors.as must be `ignore`, `notFound` or `pending`, got \"skip\"",
		},
		{
			name: "notFound without conditions",
			op:   Operation{ID: "FooGet", Type: OperationRead, Errors: []ErrorRule{{As: ErrorAsNotFound}}},
			err:  "FooGet: errors.as notFound requires status or contains",
		},
		{
			name: "invalid status",
			op:   Operation{ID: "FooGet", Type: OperationRead, Errors: []ErrorRule{{Status: 999, As: ErrorAsNotFound}}},
			err:  "FooGet: errors.status 999 is not an HTTP status",
		},
		{
			name: "pagination",
			op:   Operation{ID: "FooList", Type: OperationRead, Pagination: &Pagination{}, Errors: []ErrorRule{{Status: 400, As: ErrorAsNotFound}}},
			err:  "FooList: errors can't be combined with pagination",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateSteps(Operations{&c.op})
			if c.err == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, c.err)
		})
	}
}

func TestErrorAsSentinels(t *testing.T) {
	for _, v := range errorAsValues {
		_, ok := errorAsSentinels[v]
		require.True(t, ok, "errorAsSentinels misses %q", v)
	}
}
