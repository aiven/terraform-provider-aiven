# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

resource "aiven_mysql" "example_mysql" {
  project                 = data.aiven_project.main.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.mysql_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  mysql_user_config {
    admin_username = var.mysql_username
    admin_password = var.mysql_password

    public_access {
      mysql = true
    }
  }
}
