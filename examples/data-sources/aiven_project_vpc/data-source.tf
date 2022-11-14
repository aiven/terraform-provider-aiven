data "aiven_project_vpc" "myvpc" {
  project    = aiven_project.myproject.project
  cloud_name = "google-europe-west1"
}

# Or
data "aiven_project_vpc" "myvpc" {
  project    = aiven_project.myproject.project
  cloud_name = "google-europe-west1"
  vpc_id     = aiven_project_vpc.vpc.id
}
