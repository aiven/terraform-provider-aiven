resource "aiven_flink_jar_application" "example" {
  project      = data.aiven_project.example.project
  service_name = "example-flink-service"
  name         = "example-app-jar"
}

resource "aiven_flink_jar_application_version" "example" {
  project        = data.aiven_project.example.project
  service_name   = aiven_flink.example.service_name
  application_id = aiven_flink_application.example.application_id
  source         = "./example.jar"
}
