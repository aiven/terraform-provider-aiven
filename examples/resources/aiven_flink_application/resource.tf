resource "aiven_flink_application" "foo" {
  project = aiven_project.foo.project
  service_name = "flink-service-1"
  name = "my-flink-app"
}

