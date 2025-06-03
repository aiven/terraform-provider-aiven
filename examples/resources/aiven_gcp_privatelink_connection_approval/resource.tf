resource "aiven_gcp_privatelink_connection_approval" "approve" {
  project         = data.aiven_project.example_project.project
  service_name    = aiven_kafka.example_kafka.service_name
  user_ip_address = "10.0.0.100"
}
