data "aiven_pg" "pg" {
  project      = data.aiven_project.pr1.project
  service_name = "my-pg1"
}

