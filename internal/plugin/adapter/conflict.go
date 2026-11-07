package adapter

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	avngen "github.com/aiven/go-client-codegen"
)

// ErrConflict is returned by Create when the resource exists, see ConflictOptions.
var ErrConflict = errors.New("already exists")

// ConflictOptions makes Create fail when the resource exists in the API, or another resource
// of the same type in this provider process has the same Keys.
// The API might create the resource and still fail, so the client's retry gets 409.
type ConflictOptions struct {
	// Parent lists the attributes of the scope the Keys are unique in, e.g. project and service_name.
	Parent []string
	// Keys lists the attributes that tell the resources in the scope apart.
	Keys []string
	// Load reads the scope and returns whether d matches a loaded item.
	// It's called once per scope, unless it fails.
	Load func(ctx context.Context, client avngen.Client, d ResourceData) (func(d ResourceData) bool, error)
}

// ConflictList returns a ConflictOptions.Load that reads all pages and matches them with match(d).
func ConflictList[T any](
	pages func(ctx context.Context, client avngen.Client, d ResourceData) func(cursor string) ([]T, *string, error),
	match func(d ResourceData) func(T) bool,
) func(ctx context.Context, client avngen.Client, d ResourceData) (func(d ResourceData) bool, error) {
	return func(ctx context.Context, client avngen.Client, d ResourceData) (func(d ResourceData) bool, error) {
		items, err := Paginate(pages(ctx, client, d))
		if err != nil {
			return nil, err
		}
		return func(d ResourceData) bool {
			return slices.ContainsFunc(items, match(d))
		}, nil
	}
}

type conflictScope struct {
	sync.Mutex
	// listed is nil until the scope is listed.
	listed func(d ResourceData) bool
	// taken overrides listed: true when claimed, false when released.
	taken map[string]bool
}

var conflictScopes sync.Map

func getConflictScope(key string) *conflictScope {
	v, _ := conflictScopes.LoadOrStore(key, &conflictScope{taken: make(map[string]bool)})
	return v.(*conflictScope)
}

func conflictKeys(typeName string, opts *ConflictOptions, d ResourceData) (scope, key string) {
	values := func(attrs []string) []string {
		result := make([]string, len(attrs))
		for i, k := range attrs {
			result[i] = fmt.Sprint(d.Get(k))
		}
		return result
	}
	return conflictScopeKey(typeName, values(opts.Parent)), conflictKey(values(opts.Keys))
}

func conflictScopeKey(typeName string, parent []string) string {
	return conflictKey(append([]string{typeName}, parent...))
}

// conflictKey quotes the values, so empty values or ones with separators don't collide.
func conflictKey(values []string) string {
	return fmt.Sprintf("%q", values)
}

// claimConflict fails with ErrConflict when the resource exists, otherwise claims its Keys.
func claimConflict(ctx context.Context, client avngen.Client, typeName string, opts *ConflictOptions, d ResourceData) error {
	scopeKey, key := conflictKeys(typeName, opts, d)
	scope := getConflictScope(scopeKey)
	scope.Lock()
	defer scope.Unlock()

	if scope.listed == nil {
		listed, err := opts.Load(ctx, client, d)
		if err != nil {
			return fmt.Errorf("checking %s for conflicts: %w", typeName, err)
		}
		scope.listed = listed
	}

	taken, ok := scope.taken[key]
	if !ok {
		taken = scope.listed(d)
	}
	if taken {
		return fmt.Errorf("%s %s: %w", typeName, key, ErrConflict)
	}
	scope.taken[key] = true
	return nil
}

func releaseConflict(typeName string, opts *ConflictOptions, d ResourceData) {
	scopeKey, key := conflictKeys(typeName, opts, d)
	forgetKey(scopeKey, key)
}

// ForgetConflict releases the Keys, e.g. after the resource is deleted outside Terraform.
// The values go in the ConflictOptions.Parent and ConflictOptions.Keys order.
func ForgetConflict(typeName string, parent []string, keys ...string) {
	forgetKey(conflictScopeKey(typeName, parent), conflictKey(keys))
}

func forgetKey(scopeKey, key string) {
	v, ok := conflictScopes.Load(scopeKey)
	if !ok {
		return
	}
	scope := v.(*conflictScope)
	scope.Lock()
	defer scope.Unlock()
	scope.taken[key] = false
}
