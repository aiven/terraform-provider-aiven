# Get Aiven project details
data "aiven_project" "my_project" {
  project = "test"
}

# Create a VPC in GCP
resource "google_compute_network" "vpc" {
  name                    = "my-vpc"
  auto_create_subnetworks = "false"
}

# Create a subnet in the GCP VPC
resource "google_compute_subnetwork" "subnet" {
  name          = "my-subnet"
  region        = "us-central1"
  ip_cidr_range = "10.0.0.0/24"
  network       = google_compute_network.vpc.self_link
}

# Create a VPC in Aiven
resource "aiven_project_vpc" "my_vpc" {
  project           = data.aiven_project.my_project.project
  cloud_name        = "google-us-central1"
  network_cidr      = "192.168.0.0/24"
}

# Create a peering connection between Aiven and GCP
resource "aiven_gcp_vpc_peering_connection" "my_peering" {
  vpc_id             = aiven_project_vpc.my_vpc.id
  gcp_project_id     = var.gcp_project_id
  peer_vpc           = google_compute_network.vpc.name
}

resource "google_compute_network_peering" "aiven_peering" {
  depends_on   = [aiven_gcp_vpc_peering_connection.my_peering]
  name         = var.gcp_project_id
  network      = google_compute_network.vpc.self_link
  peer_network = aiven_gcp_vpc_peering_connection.my_peering.self_link
}