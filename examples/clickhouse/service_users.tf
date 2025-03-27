# ETL user with write permissions to the IoT measurements DB
resource "aiven_clickhouse_user" "etl" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  username     = "etl"
}

# Writer role that will be granted insert privilege to the measurements DB
resource "aiven_clickhouse_role" "writer" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  role         = "writer"
}

# Writer role's privileges
resource "aiven_clickhouse_grant" "writer_role" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  role         = aiven_clickhouse_role.writer.role

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.iot_analytics.name
  }

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.iot_analytics.name
  }
}

# Grant the writer role to the ETL user
resource "aiven_clickhouse_grant" "etl_user" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  user         = aiven_clickhouse_user.etl.username

  role_grant {
    role = aiven_clickhouse_role.writer.role
  }
}

# Analyst user with read-only access to the IoT measurements DB
resource "aiven_clickhouse_user" "analyst" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  username     = "analyst"
}

# Reader role that will be granted insert privilege to the measurements DB
resource "aiven_clickhouse_role" "reader" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  role         = "reader"
}

# Reader role's privileges
resource "aiven_clickhouse_grant" "reader_role" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  role         = aiven_clickhouse_role.reader.role

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.iot_analytics.name
  }
}

# Grant the reader role to the Analyst user
resource "aiven_clickhouse_grant" "analyst_user" {
  project      = data.aiven_project.main.project
  service_name = aiven_clickhouse.dev.service_name
  user         = aiven_clickhouse_user.analyst.username

  role_grant {
    role = aiven_clickhouse_role.reader.role
  }
}
