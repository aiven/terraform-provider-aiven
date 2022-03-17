package chsql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateRoleStatement(t *testing.T) {
	t.Run("all is well", func(tt *testing.T) {
		stmt, err := CreateRoleStatement("writer")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "CREATE ROLE IF NOT EXISTS 'writer'")
	})
	t.Run("sql is escaped", func(tt *testing.T) {
		stmt, err := CreateRoleStatement("writer; SELECT * FROM somewhere")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "CREATE ROLE IF NOT EXISTS 'writer; SELECT * FROM somewhere'")
	})
}

func TestDropRoleStatement(t *testing.T) {
	t.Run("all is well", func(tt *testing.T) {
		stmt, err := DropRoleStatement("writer")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "DROP ROLE IF EXISTS 'writer'")
	})
	t.Run("sql is escaped", func(tt *testing.T) {
		stmt, err := DropRoleStatement("writer; SELECT * FROM somewhere")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "DROP ROLE IF EXISTS 'writer; SELECT * FROM somewhere'")
	})
}

func TestShowCreateRoleStatement(t *testing.T) {
	t.Run("all is well", func(tt *testing.T) {
		stmt, err := ShowCreateRoleStatement("writer")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "SHOW CREATE ROLE 'writer'")
	})
	t.Run("sql is escaped", func(tt *testing.T) {
		stmt, err := ShowCreateRoleStatement("writer; SELECT * FROM somewhere")
		require.NoError(tt, err)
		require.Equal(tt, stmt, "SHOW CREATE ROLE 'writer; SELECT * FROM somewhere'")
	})
}
