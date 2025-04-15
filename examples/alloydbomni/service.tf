# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# AlloyDB Omni service
resource "aiven_alloydbomni" "example_alloydb" {
  project                 = data.aiven_project.main.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = var.alloydb_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  alloydbomni_user_config {
    google_columnar_engine_enabled                = true
    google_columnar_engine_memory_size_percentage = "15"
  }
}
