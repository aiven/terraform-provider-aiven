// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package clickhouse

import (
	"strings"

	"github.com/aiven/aiven-go-client"
)

// utility to check if the database returned an error because of an unknown role
// to make deletions idempotent
func IsUnknownRole(err error) bool {
	aivenError, ok := err.(aiven.Error)
	if !ok {
		return false
	}
	return strings.Contains(aivenError.Message, "Code: 511")
}
