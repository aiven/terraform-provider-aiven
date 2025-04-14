data "aiven_redis" "redis1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-redis1"
}
