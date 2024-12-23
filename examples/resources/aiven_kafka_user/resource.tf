resource "aiven_kafka_user" "example_service_user" {
  service_name = aiven_kafka.example_kafka.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-kafka-user"
  password     = var.service_user_pw
}
