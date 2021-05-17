# Cassandra service
resource "aiven_casandra" "cassandra-svc" {
  project = aiven_project.project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "samplecassandra"
}
# Forked service
resource "aiven_casandra" "cassandra-fork" {
  project = aiven_project.project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-8"
  service_name = "forkedcassandra"
  depends_on = ["aiven_casandra.cassandra-svc"]
  cassandra_user_config {
    service_to_fork_from = aiven_casandra.cassandra-svc.service_name
  }
}