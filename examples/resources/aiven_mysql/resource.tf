resource "aiven_mysql" "example_mysql" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "example-mysql"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  mysql_user_config {
    mysql_version = 8

    mysql {
      sql_mode                = "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE"
      sql_require_primary_key = true
    }

    public_access {
      mysql = true
    }
  }
}
