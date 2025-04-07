# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# Valkey service
resource "aiven_valkey" "example_valkey" {
    project      = data.aiven_project.main.project
    plan         = "startup-4"
    cloud_name   = "google-europe-west1"
    service_name = var.valkey_service_name

    valkey_user_config {
      valkey_maxmemory_policy = "allkeys-random"
    }
}