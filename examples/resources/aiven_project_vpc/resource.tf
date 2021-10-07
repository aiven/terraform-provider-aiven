resource "aiven_project_vpc" "myvpc" {
    project = aiven_project.myproject.project
    cloud_name = "google-europe-west1"
    network_cidr = "192.168.0.1/24"

    timeouts {
        create = "5m"
    }
}
