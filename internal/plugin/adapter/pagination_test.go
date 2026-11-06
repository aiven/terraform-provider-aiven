package adapter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	t.Parallel()

	type page struct {
		Names []string
		Next  *string
	}
	ptr := func(s string) *string { return &s }

	cases := []struct {
		name    string
		api     map[string]*page // pages by cursor; a missing cursor fails
		want    []string         // names of all pages
		wantErr string
	}{
		{
			name: "single page",
			api:  map[string]*page{"": {Names: []string{"a"}}},
			want: []string{"a"},
		},
		{
			name: "follows next",
			api: map[string]*page{
				"":   {Names: []string{"a"}, Next: ptr("c1")},
				"c1": {Names: []string{"b"}, Next: ptr("c2")},
				"c2": {Names: []string{"c"}},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "empty page isn't nil",
			api:  map[string]*page{"": {Names: []string{}}},
			want: []string{},
		},
		{
			name: "empty next stops",
			api:  map[string]*page{"": {Names: []string{"a"}, Next: ptr("")}},
			want: []string{"a"},
		},
		{
			name:    "error",
			api:     map[string]*page{"": {Names: []string{"a"}, Next: ptr("c1")}},
			wantErr: `no page "c1"`,
		},
		{
			name: "repeated cursor",
			api: map[string]*page{
				"":   {Names: []string{"a"}, Next: ptr("c1")},
				"c1": {Names: []string{"b"}, Next: ptr("c1")},
			},
			wantErr: `pagination cursor "c1" repeated`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Paginate(func(cursor string) ([]string, *string, error) {
				p, ok := tc.api[cursor]
				if !ok {
					return nil, nil, fmt.Errorf("no page %q", cursor)
				}
				return p.Names, p.Next, nil
			})
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestQueryParams(t *testing.T) {
	t.Parallel()

	require.Equal(t, [][2]string{{"limit", "10"}}, QueryParams("cursor", "", "limit", "10"))
	require.Equal(t, [][2]string{{"cursor", "c1"}, {"limit", "10"}}, QueryParams("cursor", "c1", "limit", "10"))
	require.Empty(t, QueryParams())
	odd := []string{"cursor"}
	require.Panics(t, func() { QueryParams(odd...) })
}

func TestPaginateFind(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }
	pages := map[string][]string{"": {"a", "b"}, "c1": {"c", "c"}, "c2": {"e"}}
	nexts := map[string]*string{"": ptr("c1"), "c1": ptr("c2")}

	cases := []struct {
		name        string
		find        string
		wantFetched []string
		wantErr     error
	}{
		{name: "first page", find: "b", wantFetched: []string{""}},
		{name: "first match of second page", find: "c", wantFetched: []string{"", "c1"}},
		{name: "not found", find: "z", wantFetched: []string{"", "c1", "c2"}, wantErr: ErrNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var fetched []string
			got, err := PaginateFind(func(cursor string) ([]string, *string, error) {
				fetched = append(fetched, cursor)
				return pages[cursor], nexts[cursor], nil
			}, func(s string) bool { return s == tc.find })
			require.Equal(t, tc.wantFetched, fetched)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.find, got)
		})
	}
}
