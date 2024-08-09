resource "aiven_azure_privatelink" "main" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_kafka.example_kafka.service_name


  user_subscription_ids = [
    "00000000-0000-0000-0000-000000000000"
  ]
}
