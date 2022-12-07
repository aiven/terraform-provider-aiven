data "aiven_pg_database" "mydatabase" {
  project       = aiven_project.myproject.project
  service_name  = aiven_pg.mypg.service_name
  database_name = "<DATABASE_NAME>"
}

