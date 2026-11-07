package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/stretchr/testify/require"
)

func conflictTestTypeName(t *testing.T) string {
	t.Helper()
	return strings.ReplaceAll(t.Name(), "/", "_")
}

func conflictTestData(t *testing.T, plan map[string]any) ResourceData {
	t.Helper()
	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"project":       {Type: SchemaTypeString},
			"service_name":  {Type: SchemaTypeString},
			"database_name": {Type: SchemaTypeString},
		},
	}
	d, err := NewResourceData(sch, []string{"project", "service_name", "database_name"}, WithTestPlan(plan))
	require.NoError(t, err)
	return d
}

func conflictTestDataNameKind(t *testing.T, plan map[string]any) ResourceData {
	t.Helper()
	sch := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"project": {Type: SchemaTypeString},
			"name":    {Type: SchemaTypeString},
			"kind":    {Type: SchemaTypeString},
		},
	}
	d, err := NewResourceData(sch, []string{"project", "name", "kind"}, WithTestPlan(plan))
	require.NoError(t, err)
	return d
}

func listedDatabaseMatcher(names ...string) func(ResourceData) bool {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(d ResourceData) bool {
		_, ok := set[fmt.Sprint(d.Get("database_name"))]
		return ok
	}
}

func TestClaimConflictListsOncePerScope(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	listCalls := 0
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			listCalls++
			return listedDatabaseMatcher("existing"), nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "new-db",
	})

	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
	d2 := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "new-db-2",
	})
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d2))
	require.Equal(t, 1, listCalls)
}

func TestClaimConflictConflictsWithListedName(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return listedDatabaseMatcher("taken"), nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "taken",
	})

	err := claimConflict(ctx, nil, typeName, opts, d)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrConflict)
}

func TestClaimConflictConflictsOnSecondClaim(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return func(ResourceData) bool { return false }, nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "db1",
	})

	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
	err := claimConflict(ctx, nil, typeName, opts, d)
	require.ErrorIs(t, err, ErrConflict)
}

func TestClaimConflictDifferentParentScopes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	listCalls := 0
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			listCalls++
			return func(ResourceData) bool { return false }, nil
		},
	}
	d1 := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "db1",
	})
	d2 := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s2", "database_name": "db1",
	})

	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d1))
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d2))
	require.Equal(t, 2, listCalls)
}

func TestClaimConflictListErrorNotCached(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	listCalls := 0
	listErr := errors.New("list failed")
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			listCalls++
			if listCalls == 1 {
				return nil, listErr
			}
			return func(ResourceData) bool { return false }, nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "db1",
	})

	err := claimConflict(ctx, nil, typeName, opts, d)
	require.ErrorContains(t, err, "checking "+typeName)
	require.ErrorIs(t, err, listErr)
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
	require.Equal(t, 2, listCalls)
}

func TestReleaseConflictAndForgetConflictAllowReclaim(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return func(ResourceData) bool { return false }, nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "db1",
	})

	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
	require.ErrorIs(t, claimConflict(ctx, nil, typeName, opts, d), ErrConflict)

	releaseConflict(typeName, opts, d)
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))

	releaseConflict(typeName, opts, d)
	ForgetConflict(typeName, []string{"p1", "s1"}, "db1")
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
}

func TestClaimConflictMultiKey(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project"},
		Keys:   []string{"name", "kind"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return func(d ResourceData) bool {
				return d.Get("name") == "foo" && d.Get("kind") == "a"
			}, nil
		},
	}
	dOtherKind := conflictTestDataNameKind(t, map[string]any{
		"project": "p1", "name": "foo", "kind": "b",
	})
	dSame := conflictTestDataNameKind(t, map[string]any{
		"project": "p1", "name": "foo", "kind": "a",
	})

	require.NoError(t, claimConflict(ctx, nil, typeName, opts, dOtherKind))
	err := claimConflict(ctx, nil, typeName, opts, dSame)
	require.ErrorIs(t, err, ErrConflict)

	require.ErrorIs(t, claimConflict(ctx, nil, typeName, opts, dOtherKind), ErrConflict)
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, conflictTestDataNameKind(t, map[string]any{
		"project": "p1", "name": "foo", "kind": "c",
	})))
}

func TestClaimConflictKeysDontCollide(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project"},
		Keys:   []string{"name", "kind"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return func(ResourceData) bool { return false }, nil
		},
	}
	for _, plan := range []map[string]any{
		{"project": "p1", "name": "a", "kind": ""},
		{"project": "p1", "name": "", "kind": "a"},
		{"project": "p1", "name": "a/b", "kind": "c"},
		{"project": "p1", "name": "a", "kind": "b/c"},
	} {
		require.NoError(t, claimConflict(ctx, nil, typeName, opts, conflictTestDataNameKind(t, plan)), plan)
	}
}

func TestReleaseConflictFreesListedKey(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	typeName := conflictTestTypeName(t)
	opts := &ConflictOptions{
		Parent: []string{"project", "service_name"},
		Keys:   []string{"database_name"},
		Load: func(context.Context, avngen.Client, ResourceData) (func(ResourceData) bool, error) {
			return listedDatabaseMatcher("taken"), nil
		},
	}
	d := conflictTestData(t, map[string]any{
		"project": "p1", "service_name": "s1", "database_name": "taken",
	})

	require.ErrorIs(t, claimConflict(ctx, nil, typeName, opts, d), ErrConflict)

	releaseConflict(typeName, opts, d)
	require.NoError(t, claimConflict(ctx, nil, typeName, opts, d))
}

func TestConflictListMergesPages(t *testing.T) {
	t.Parallel()

	type row struct{ Name string }
	next := func(s string) *string { return &s }
	pages := func(_ context.Context, _ avngen.Client, _ ResourceData) func(cursor string) ([]row, *string, error) {
		return func(cursor string) ([]row, *string, error) {
			switch cursor {
			case "":
				return []row{{"a"}, {"b"}}, next("c1"), nil
			case "c1":
				return []row{{"c"}}, nil, nil
			default:
				return nil, nil, fmt.Errorf("unknown cursor %q", cursor)
			}
		}
	}
	list := ConflictList(pages, func(d ResourceData) func(row) bool {
		return func(item row) bool {
			return item.Name == fmt.Sprint(d.Get("database_name"))
		}
	})

	matches, err := list(t.Context(), nil, conflictTestData(t, map[string]any{
		"project": "p", "service_name": "s", "database_name": "b",
	}))
	require.NoError(t, err)
	require.True(t, matches(conflictTestData(t, map[string]any{
		"project": "p", "service_name": "s", "database_name": "a",
	})))
	require.True(t, matches(conflictTestData(t, map[string]any{
		"project": "p", "service_name": "s", "database_name": "b",
	})))
	require.True(t, matches(conflictTestData(t, map[string]any{
		"project": "p", "service_name": "s", "database_name": "c",
	})))
	require.False(t, matches(conflictTestData(t, map[string]any{
		"project": "p", "service_name": "s", "database_name": "missing",
	})))
}
