resource "aiven_pg_user" "example_user" {
  service_name = aiven_pg.example_postgres.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-service-user"
  password     = var.service_user_password
}

# Each service has a default admin user with the username avnadmin.
resource "aiven_pg_user" "admin_user" {
  service_name         = aiven_pg.example_postgres.service_name
  project              = data.aiven_project.example_project.project
  username             = "avnadmin"
  password             = var.service_user_password
  # The pg_allow_replication attribute is required for this user and must be true.
  pg_allow_replication = true
}