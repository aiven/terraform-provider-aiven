data "aiven_project_vpc" "example_vpc" {
  project    = data.aiven_project.example_project.project
  cloud_name = "google-europe-west1"
}
