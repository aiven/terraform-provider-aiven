resource "aiven_project" "clickhouse_dev" {
  project = "clickhouse-dev"
}

# ClickHouse service in the same region
resource "aiven_clickhouse" "dev" {
  project                 = aiven_project.clickhouse_dev.project
  cloud_name              = "google-europe-west1"
  plan                    = "hobbyist"
  service_name            = "clickhouse-gcp-eu"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

# Sample ClickHouse database that can be used to write the raw data
resource "aiven_clickhouse_database" "iot_analytics" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  name         = "iot_analytics"
}

# ETL user with write permissions to the IoT measurements DB
resource "aiven_clickhouse_user" "etl" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  username     = "etl"
}

# Writer role that will be granted insert privilege to the measurements DB
resource "aiven_clickhouse_role" "writer" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  role         = "writer"
}

# Writer role's privileges
resource "aiven_clickhouse_grant" "writer_role" {
  project      = aiven_project.clickhouse_dev.project
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
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  user         = aiven_clickhouse_user.etl.username

  role_grant {
    role = aiven_clickhouse_role.writer.role
  }
}

# Analyst user with read-only access to the IoT measurements DB
resource "aiven_clickhouse_user" "analyst" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  username     = "analyst"
}

# Reader role that will be granted insert privilege to the measurements DB
resource "aiven_clickhouse_role" "reader" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  role         = "reader"
}

# Reader role's privileges
resource "aiven_clickhouse_grant" "reader_role" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  role         = aiven_clickhouse_role.reader.role

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.iot_analytics.name
  }
}

# Grant the reader role to the Analyst user
resource "aiven_clickhouse_grant" "analyst_user" {
  project      = aiven_project.clickhouse_dev.project
  service_name = aiven_clickhouse.dev.service_name
  user         = aiven_clickhouse_user.analyst.username

  role_grant {
    role = aiven_clickhouse_role.reader.role
  }
}
