# Cassandra service
resource "aiven_service" "cassandra-svc" {
  project = aiven_project.project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "samplecassandra"
  service_type = "cassandra"
}
# Forked service
resource "aiven_service" "cassandra-fork" {
  project = aiven_project.project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-8"
  service_name = "forkedcassandra"
  service_type = "cassandra"
  depends_on = ["aiven_service.cassandra-svc"]
  cassandra_user_config {
    service_to_fork_from = aiven_service.cassandra-svc.service_name
  }
}