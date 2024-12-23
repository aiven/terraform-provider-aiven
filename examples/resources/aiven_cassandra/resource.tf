resource "aiven_cassandra" "example_cassandra" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "example-cassandra-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  cassandra_user_config {
    migrate_sstableloader = true

    public_access {
      prometheus = true
    }
  }
}
