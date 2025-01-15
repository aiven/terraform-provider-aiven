resource "aiven_flink_jar_application" "example" {
  project      = data.aiven_project.example.project
  service_name = "example-flink-service"
  name         = "example-app-jar"
}
