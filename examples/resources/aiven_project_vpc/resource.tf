resource "aiven_project_vpc" "example_vpc" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "google-europe-west1"
  network_cidr = "192.168.1.0/24"

  timeouts {
    create = "5m"
  }
}
