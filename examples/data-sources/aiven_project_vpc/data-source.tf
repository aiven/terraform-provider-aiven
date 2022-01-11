data "aiven_project_vpc" "myvpc" {
  project    = aiven_project.myproject.project
  cloud_name = "google-europe-west1"
}

