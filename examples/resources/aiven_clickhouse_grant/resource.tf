resource "aiven_clickhouse_role" "example_role" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  role         = "example-role"
}

# Grant privileges to the example role.
resource "aiven_clickhouse_grant" "role_privileges" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  role         = aiven_clickhouse_role.example_role.role

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.example_db.name
    table     = "example-table"
  }

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.example_db.name
  }

  # Global privileges
    privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }

  privilege_grant {
    privilege = "SYSTEM DROP CACHE"
    database  = "*"
  }
}

# Grant the role to the user.
resource "aiven_clickhouse_user" "example_user" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  username     = "example-user"
}

resource "aiven_clickhouse_grant" "user_role_assignment" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  user         = aiven_clickhouse_user.example_user.username

  role_grant {
    role = aiven_clickhouse_role.example_role.role
  }
}
