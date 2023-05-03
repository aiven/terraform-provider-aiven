resource "aiven_clickhouse" "clickhouse" {
  project      = var.aiven_project_name
  cloud_name   = "google-europe-west1"
  plan         = "startup-8"
  service_name = "exapmle-clickhouse"
}

resource "aiven_clickhouse_database" "demodb" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  name         = "demo"
}

resource "aiven_clickhouse_role" "demo" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  role         = "demo-role"
}

resource "aiven_clickhouse_grant" "demo-role-grant" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  role         = aiven_clickhouse_role.demo.role

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.demodb.name
    table     = "demo-table"
  }

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.demodb.name
  }
}

resource "aiven_clickhouse_user" "demo" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  username     = "demo-user"
}

resource "aiven_clickhouse_grant" "demo-user-grant" {
  project      = aiven_clickhouse.clickhouse.project
  service_name = aiven_clickhouse.clickhouse.service_name
  user         = aiven_clickhouse_user.demo.username

  role_grant {
    role = aiven_clickhouse_role.demo.role
  }
}
