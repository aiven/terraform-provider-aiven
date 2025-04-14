data "aiven_azure_privatelink" "main" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_kafka.example_kafka.service_name
}
