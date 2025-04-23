# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# Dragonfly service
resource "aiven_dragonfly" "example_dragonfly" {
    project      = data.aiven_project.main.project
    plan         = "startup-4"
    cloud_name   = "google-europe-west1"
    service_name = var.dragonfly_service_name

    dragonfly_user_config {
        cache_mode = true
    }
}
