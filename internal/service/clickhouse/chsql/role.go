package chsql

import (
	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse/chsql/sanitize"
)

const (
	// DefaultDatabaseForRoles this will be the default database to assume when non is present in a statement
	// just use one that always exists
	DefaultDatabaseForRoles = "system"
)

func CreateRoleStatement(roleName string) (string, error) {
	return sanitize.SanitizeSQL("CREATE ROLE IF NOT EXISTS $1", roleName)
}

func DropRoleStatement(roleName string) (string, error) {
	return sanitize.SanitizeSQL("DROP ROLE IF EXISTS $1", roleName)
}

func ShowCreateRoleStatement(roleName string) (string, error) {
	return sanitize.SanitizeSQL("SHOW CREATE ROLE $1", roleName)
}
