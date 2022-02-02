data "aiven_clickhouse_user" "ch-user" {
  project      = aiven_project.myproject.project
  service_name = aiven_clickhouse.myservice.service_name
  username     = "<USERNAME>"
}
