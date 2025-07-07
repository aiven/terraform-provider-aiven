# Manage multiple ClickHouse grants for a user or role

This example demonstrates how to properly manage ClickHouse grants in complex scenarios, combining global and database-specific privileges while avoiding overlaps that could cause conflicts.

## Overlapping grants

You can grant privileges to each role or user using **only one** `aiven_clickhouse_grant` resource. Multiple grant resources for the same entity overlap and overwrite each other.

For example, the following file has two `aiven_clickhouse_grant` resources for the same `dba` role, which cause the role to be overwritten:
```hcl
resource "aiven_clickhouse_grant" "dba_global" {
  role = "dba"
  privilege_grant {
    privilege = "SHOW"
    database  = "*"
  }
}

resource "aiven_clickhouse_grant" "dba_specific" {
  role = "dba"  # Same role - will overwrite the first grant!
  privilege_grant {
    privilege = "SELECT"
    database  = "main_db"
  }
}
```

To avoid the overlap, use a single grant resource to grant these privileges to the role:
```hcl
resource "aiven_clickhouse_grant" "dba_combined" {
  role = "dba"
  # All privileges for this role in one resource
  privilege_grant {
    privilege = "SHOW"
    database  = "*"
  }
  privilege_grant {
    privilege = "SELECT"
    database  = "main_db"
  }
}
```

## Example solution

This example demonstrates best practices for granting privileges:

1. **Single grant resource**: Use one `aiven_clickhouse_grant` for each role and user.
2. **Dynamic privilege blocks**: Use `dynamic "privilege_grant"` to combine all privileges.
3. **Locals for organization**: Separate global and specific grants for clarity.
4. **Proper scoping**: Grant privileges at only the level needed by specifying a database or setting privileges globally using `database = "*"`.
