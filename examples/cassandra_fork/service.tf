# Cassandra service
resource "aiven_cassandra" "cassandra-svc" {
  project      = aiven_project.project1.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "samplecassandra"
}
# Forked service
resource "aiven_cassandra" "cassandra-fork" {
  project      = aiven_project.project1.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-8"
  service_name = "forkedcassandra"
  cassandra_user_config {
    service_to_fork_from = aiven_cassandra.cassandra-svc.service_name
  }
}
