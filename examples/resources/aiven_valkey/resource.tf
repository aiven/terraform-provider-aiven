resource "aiven_valkey" "example_valkey" {
  project      = data.aiven_project.example_project.project
  plan         = "startup-4"
  cloud_name   = "google-europe-west1"
  service_name = "example-valkey-service"

  valkey_user_config {
    valkey_maxmemory_policy = "allkeys-random"
  }
}
