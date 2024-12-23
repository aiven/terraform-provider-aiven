# Create an ETL user
resource "aiven_clickhouse_user" "etl" {
  project      = aiven_project.clickhouse_kafka_source.project
  service_name = aiven_clickhouse.clickhouse.service_name
  username     = "etl"
}

# Create a role named writer
resource "aiven_clickhouse_role" "writer" {
  project      = aiven_project.clickhouse_kafka_source.project
  service_name = aiven_clickhouse.clickhouse.service_name
  role         = "writer"
}

# Set the privileges for the writer role
resource "aiven_clickhouse_grant" "writer_role" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
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
# to give them write permissions to the IoT measurements database
resource "aiven_clickhouse_grant" "etl_user" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  user         = aiven_clickhouse_user.etl.username

  role_grant {
    role = aiven_clickhouse_role.writer.role
  }
}

# Create an analyst
resource "aiven_clickhouse_user" "analyst" {
  project      = aiven_project.clickhouse_kafka_source.project
  service_name = aiven_clickhouse.clickhouse.service_name
  username     = "analyst"
}

# Create a role named reader role
resource "aiven_clickhouse_role" "reader" {
  project      = aiven_project.clickhouse_kafka_source.project
  service_name = aiven_clickhouse.clickhouse.service_name
  role         = "reader"
}

# Set the privileges for the reader role
resource "aiven_clickhouse_grant" "reader_role" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  role         = aiven_clickhouse_role.reader.role

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.iot_analytics.name
  }
}

# Grant the reader role to the analyst user
# to give them insert privileges to the measurements database
resource "aiven_clickhouse_grant" "analyst_user" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  user         = aiven_clickhouse_user.analyst.username

  role_grant {
    role = aiven_clickhouse_role.reader.role
  }
}
