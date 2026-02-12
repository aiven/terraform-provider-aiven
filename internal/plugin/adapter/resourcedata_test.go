// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestDereference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		value           any
		want            any
		wantErrorSubstr string
	}{
		{name: "nil", value: nil, want: nil},
		{name: "string", value: "hello", want: "hello"},
		{name: "int", value: 42, want: 42},
		{name: "bool", value: true, want: true},
		{name: "float64", value: 3.14, want: 3.14},
		{name: "slice", value: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "map", value: map[string]int{"x": 1}, want: map[string]int{"x": 1}},
		{name: "pointer to string", value: lo.ToPtr("hello"), want: "hello"},
		{name: "pointer to int", value: lo.ToPtr(42), want: 42},
		{name: "pointer to bool", value: lo.ToPtr(true), want: true},
		{name: "pointer to float64", value: lo.ToPtr(3.14), want: 3.14},
		{name: "pointer to slice", value: lo.ToPtr([]string{"a", "b"}), want: []string{"a", "b"}},
		{name: "pointer to map", value: lo.ToPtr(map[string]int{"x": 1}), want: map[string]int{"x": 1}},
		{name: "pointer to pointer to string", value: lo.ToPtr(lo.ToPtr("hello")), wantErrorSubstr: "pointer to pointer"},
		{name: "3 pointers to string", value: lo.ToPtr(lo.ToPtr(lo.ToPtr("hello"))), wantErrorSubstr: "pointer to pointer"},
		{name: "nil pointer", value: (*string)(nil), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dereference(tt.value)
			if tt.wantErrorSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrorSubstr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
