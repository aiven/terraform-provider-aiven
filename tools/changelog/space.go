package changelog

import "strings"

// NormalizeSpace collapses whitespace runs to single spaces.
func NormalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
