package adapter

import (
	"fmt"
	"slices"
)

// Paginate calls fetch with the cursor it returned last until it's nil or "", and returns the items of all pages.
func Paginate[T any](fetch func(cursor string) ([]T, *string, error)) ([]T, error) {
	var all []T
	err := eachPage(fetch, func(items []T) bool {
		// Keeps an empty first page non-nil, so it flattens to [], not null
		if all == nil {
			all = items
		} else {
			all = append(all, items...)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return all, nil
}

// PaginateFind returns the first item for which found is true, and fetches no pages after it.
// Returns ErrNotFound when no page has it.
func PaginateFind[T any](fetch func(cursor string) ([]T, *string, error), found func(T) bool) (T, error) {
	var match T
	ok := false
	err := eachPage(fetch, func(items []T) bool {
		if i := slices.IndexFunc(items, found); i >= 0 {
			match, ok = items[i], true
		}
		return ok
	})
	if err == nil && !ok {
		err = ErrNotFound
	}
	return match, err
}

// eachPage passes the items of every page to page until it returns true.
// A repeated cursor fails, so the API can't loop forever.
func eachPage[T any](fetch func(cursor string) ([]T, *string, error), page func([]T) bool) error {
	cursor := ""
	seen := make(map[string]bool)
	for {
		items, next, err := fetch(cursor)
		if err != nil {
			return err
		}
		if page(items) || next == nil || *next == "" {
			return nil
		}
		if seen[*next] {
			return fmt.Errorf("pagination cursor %q repeated", *next)
		}
		seen[*next] = true
		cursor = *next
	}
}

// QueryParams packs key-value pairs into client query params, skipping empty values, e.g. the first page cursor.
func QueryParams(kv ...string) [][2]string {
	if len(kv)%2 != 0 {
		panic(fmt.Errorf("query params: odd number of arguments %q", kv))
	}
	query := make([][2]string, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		if kv[i+1] != "" {
			query = append(query, [2]string{kv[i], kv[i+1]})
		}
	}
	return query
}
