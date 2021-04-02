#Mysql database creation
resource "aiven_mysql" "mysql" {
  project                 = var.aiven_project_id
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "mysql"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  mysql_user_config {
    admin_password = "aiven"
    admin_username = "aiven@123"
    mysql_version  = 8

    public_access {
      mysql = true
    }
  }
}